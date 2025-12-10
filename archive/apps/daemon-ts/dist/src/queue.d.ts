import { Logger } from "./logger.js";
export declare class BuildQueue {
    private semaphore;
    private waiting;
    private logger;
    constructor(maxConcurrent: number, logger: Logger);
    acquire(): Promise<void>;
    release(): void;
}
