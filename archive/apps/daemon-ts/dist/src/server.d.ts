import { Logger } from "./logger.js";
import { ConfigManager } from "./config.js";
import { DB } from "./database.js";
import { DockerManager } from "./docker.js";
export interface BuildInfo {
    id: string;
    target: string;
    state: string;
    startTime: number;
    endTime?: number;
    exitCode?: number;
    errorMessage?: string;
    configPath: string;
    customer?: string;
    logs: LogEntry[];
    logSubscribers: Set<(log: LogEntry) => void>;
    artifactPaths: string[];
    cancel?: () => void;
}
export interface LogEntry {
    timestampUnixSeconds: number;
    stream: string;
    message: string;
}
export declare class SmidrServer {
    private grpcServer;
    private address;
    private port;
    private logger;
    private configManager;
    private database;
    private dockerManager;
    private buildQueue;
    private builds;
    private customerQueues;
    constructor(address: string | undefined, port: number | undefined, logger: Logger, configManager: ConfigManager, database: DB, dockerManager: DockerManager);
    start(): Promise<void>;
    stop(): Promise<void>;
    private startBuild;
    private executeBuild;
    private getBuildStatus;
    private listBuilds;
    private cancelBuild;
    private streamBuildLogs;
    private listArtifacts;
}
