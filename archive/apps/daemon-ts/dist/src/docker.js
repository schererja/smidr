import Docker from "dockerode";
export class DockerManager {
    constructor(logger) {
        this.docker = new Docker();
        this.logger = logger;
    }
    async runBuild(buildId, config, target, logCallback) {
        try {
            this.logger.info(`Running build ${buildId} for target ${target}`);
            // Create container
            const container = await this.docker.createContainer({
                Image: "yoctoproject/poky:latest", // Example image
                Cmd: ["bitbake", target],
                Tty: true,
                Volumes: {
                    "/build": {},
                },
                Env: [`BB_NUMBER_THREADS=4`, `PARALLEL_MAKE=-j4`],
            });
            // Attach to logs
            const stream = await container.attach({
                stream: true,
                stdout: true,
                stderr: true,
            });
            stream.on("data", (chunk) => {
                logCallback("stdout", chunk.toString());
            });
            stream.on("error", (error) => {
                logCallback("stderr", error.message);
            });
            // Start container
            await container.start();
            // Wait for container to finish
            const exitInfo = await container.wait();
            // Clean up
            await container.remove();
            return {
                exitCode: exitInfo.StatusCode,
                artifacts: [],
            };
        }
        catch (error) {
            this.logger.error(`Build ${buildId} failed`, error);
            return {
                exitCode: 1,
                artifacts: [],
                error: error.message,
            };
        }
    }
    async isRunning() {
        try {
            await this.docker.ping();
            return true;
        }
        catch {
            return false;
        }
    }
}
