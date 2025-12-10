import * as grpc from "@grpc/grpc-js";
import * as protoLoader from "@grpc/proto-loader";
import { v4 as uuidv4 } from "uuid";
import path from "path";
import { fileURLToPath } from "url";
import { BuildQueue } from "./queue.js";
const __dirname = path.dirname(fileURLToPath(import.meta.url));
export class SmidrServer {
    constructor(address = "0.0.0.0", port = 50051, logger, configManager, database, dockerManager) {
        this.builds = new Map();
        this.customerQueues = new Map();
        this.address = address;
        this.port = port;
        this.logger = logger;
        this.configManager = configManager;
        this.database = database;
        this.dockerManager = dockerManager;
        this.grpcServer = new grpc.Server();
        this.buildQueue = new BuildQueue(5, logger); // Global semaphore of 5 concurrent builds
    }
    async start() {
        try {
            // Load proto definitions from project root
            const protoDir = process.env.PROTO_DIR ||
                path.resolve(process.cwd(), '../..', 'protos/smidr/v1');
            const protoPath = path.join(protoDir, 'smidr_service.proto');
            const packageDef = await protoLoader.load(protoPath, {
                keepCase: true,
                longs: String,
                enums: String,
                defaults: true,
                oneofs: true,
            });
            const proto = grpc.loadPackageDefinition(packageDef);
            // Register service implementations
            this.grpcServer.addService(proto.smidr.v1.BuildService.service, {
                startBuild: this.startBuild.bind(this),
                getBuildStatus: this.getBuildStatus.bind(this),
                listBuilds: this.listBuilds.bind(this),
                cancelBuild: this.cancelBuild.bind(this),
            });
            this.grpcServer.addService(proto.smidr.v1.LogService.service, {
                streamBuildLogs: this.streamBuildLogs.bind(this),
            });
            this.grpcServer.addService(proto.smidr.v1.ArtifactService.service, {
                listArtifacts: this.listArtifacts.bind(this),
            });
            // Start server
            await new Promise((resolve, reject) => {
                this.grpcServer.bindAsync(`${this.address}:${this.port}`, grpc.ServerCredentials.createInsecure(), (err) => {
                    if (err)
                        reject(err);
                    else
                        resolve();
                });
            });
            this.logger.info(`gRPC server listening on ${this.address}:${this.port}`);
        }
        catch (error) {
            this.logger.error("Failed to start server", error);
            throw error;
        }
    }
    async stop() {
        return new Promise((resolve) => {
            this.grpcServer.tryShutdown(() => {
                this.logger.info("gRPC server shut down");
                resolve();
            });
        });
    }
    async startBuild(call, callback) {
        try {
            const { config_path, target, customer, force_clean, force_image_rebuild, } = call.request;
            const buildId = `${customer || "build"}-${uuidv4().slice(0, 8)}`;
            // Load config
            const config = await this.configManager.loadConfig(config_path);
            // Create build info
            const buildInfo = {
                id: buildId,
                target,
                state: "BUILDING",
                startTime: Math.floor(Date.now() / 1000),
                customer,
                configPath: config_path,
                logs: [],
                logSubscribers: new Set(),
                artifactPaths: [],
            };
            this.builds.set(buildId, buildInfo);
            // Queue the build
            const customerKey = customer || "default";
            if (!this.customerQueues.has(customerKey)) {
                this.customerQueues.set(customerKey, new BuildQueue(1, this.logger));
            }
            const customerQueue = this.customerQueues.get(customerKey);
            // Start build in background
            this.executeBuild(buildInfo, config, force_clean, force_image_rebuild, customerQueue);
            callback(null, {
                build_identifier: { build_id: buildId },
                state: "BUILDING",
                timestamps: { start_time_unix_seconds: buildInfo.startTime },
            });
        }
        catch (error) {
            this.logger.error("Failed to start build", error);
            callback(error);
        }
    }
    async executeBuild(buildInfo, config, forceClean, forceImageRebuild, queue) {
        try {
            // Acquire semaphore slot
            await this.buildQueue.acquire();
            this.logger.info(`Starting build ${buildInfo.id}`);
            // Run in Docker
            const result = await this.dockerManager.runBuild(buildInfo.id, config, buildInfo.target, (stream, message) => {
                const logEntry = {
                    timestampUnixSeconds: Math.floor(Date.now() / 1000),
                    stream,
                    message,
                };
                buildInfo.logs.push(logEntry);
                buildInfo.logSubscribers.forEach((callback) => callback(logEntry));
            });
            buildInfo.state = result.exitCode === 0 ? "COMPLETED" : "FAILED";
            buildInfo.exitCode = result.exitCode;
            buildInfo.endTime = Math.floor(Date.now() / 1000);
            buildInfo.artifactPaths = result.artifacts;
            if (result.exitCode !== 0) {
                buildInfo.errorMessage = result.error;
            }
            // Persist to database
            this.database.saveBuild(buildInfo.id, buildInfo.target, buildInfo.state, buildInfo.startTime, buildInfo.customer, buildInfo.configPath);
            this.logger.info(`Build ${buildInfo.id} completed with status ${buildInfo.state}`);
        }
        catch (error) {
            buildInfo.state = "FAILED";
            buildInfo.errorMessage = error.message;
            buildInfo.endTime = Math.floor(Date.now() / 1000);
            this.logger.error(`Build ${buildInfo.id} failed`, error);
        }
        finally {
            this.buildQueue.release();
        }
    }
    async getBuildStatus(call, callback) {
        try {
            const { build_identifier } = call.request;
            const buildId = build_identifier?.build_id;
            const buildInfo = this.builds.get(buildId);
            if (!buildInfo) {
                throw new Error(`Build not found: ${buildId}`);
            }
            callback(null, {
                build_identifier: { build_id: buildInfo.id },
                state: buildInfo.state,
                timestamps: {
                    start_time_unix_seconds: buildInfo.startTime,
                    end_time_unix_seconds: buildInfo.endTime || 0,
                },
                exit_code: buildInfo.exitCode || 0,
                error_message: buildInfo.errorMessage,
            });
        }
        catch (error) {
            callback(error);
        }
    }
    async listBuilds(call, callback) {
        try {
            const builds = Array.from(this.builds.values()).map((b) => ({
                build_identifier: { build_id: b.id },
                target_image: b.target,
                build_state: b.state,
                timestamps: {
                    start_time_unix_seconds: b.startTime,
                    end_time_unix_seconds: b.endTime || 0,
                },
                customer: b.customer,
            }));
            callback(null, { builds });
        }
        catch (error) {
            callback(error);
        }
    }
    async cancelBuild(call, callback) {
        try {
            const { build_identifier } = call.request;
            const buildId = build_identifier?.build_id;
            const buildInfo = this.builds.get(buildId);
            if (!buildInfo) {
                throw new Error(`Build not found: ${buildId}`);
            }
            if (buildInfo.cancel) {
                buildInfo.cancel();
            }
            buildInfo.state = "CANCELLED";
            buildInfo.endTime = Math.floor(Date.now() / 1000);
            callback(null, { success: true, message: `Build ${buildId} cancelled` });
        }
        catch (error) {
            callback(error);
        }
    }
    streamBuildLogs(call) {
        try {
            const { build_identifier } = call.request;
            const buildId = build_identifier?.build_id;
            const buildInfo = this.builds.get(buildId);
            if (!buildInfo) {
                call.destroy(new Error(`Build not found: ${buildId}`));
                return;
            }
            // Send existing logs
            buildInfo.logs.forEach((log) => {
                call.write(log);
            });
            // Subscribe to new logs
            const callback = (log) => {
                call.write(log);
            };
            buildInfo.logSubscribers.add(callback);
            call.on("end", () => {
                buildInfo.logSubscribers.delete(callback);
                call.end();
            });
        }
        catch (error) {
            call.destroy(error);
        }
    }
    async listArtifacts(call, callback) {
        try {
            const { build_identifier } = call.request;
            const buildId = build_identifier?.build_id;
            const buildInfo = this.builds.get(buildId);
            if (!buildInfo) {
                throw new Error(`Build not found: ${buildId}`);
            }
            const artifacts = buildInfo.artifactPaths.map((path) => ({
                name: path,
                size_bytes: 0,
                download_url: `/artifacts/${buildId}/${path}`,
                checksum: "",
            }));
            callback(null, { artifacts });
        }
        catch (error) {
            callback(error);
        }
    }
}
