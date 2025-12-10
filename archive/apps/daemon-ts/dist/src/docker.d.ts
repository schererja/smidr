import { Logger } from "./logger.js";
export declare class DockerManager {
    private docker;
    private logger;
    constructor(logger: Logger);
    runBuild(buildId: string, config: any, target: string, logCallback: (stream: string, message: string) => void): Promise<{
        exitCode: number;
        artifacts: string[];
        error?: string;
    }>;
    isRunning(): Promise<boolean>;
}
