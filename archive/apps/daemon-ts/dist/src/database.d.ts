interface Build {
    id: string;
    customer_id: string;
    target: string;
    config_path: string;
    state: string;
    start_time: number;
    end_time?: number;
    exit_code?: number;
}
interface LogEntry {
    timestamp: number;
    stream: "stdout" | "stderr";
    message: string;
}
export declare class DB {
    private db;
    private dbPath;
    private isInitialized;
    constructor(dbPath?: string);
    initialize(): Promise<void>;
    private execute;
    private query;
    saveBuild(id: string, target: string, state: string, startTime: number, customerId?: string, configPath?: string): void;
    getBuild(buildId: string): Build | null;
    listBuilds(customerId: string, limit?: number): Build[];
    updateBuildState(buildId: string, state: string, endTime?: number, exitCode?: number): void;
    saveLogs(buildId: string, logs: LogEntry[]): void;
    getLogs(buildId: string): LogEntry[];
    private saveToFile;
    close(): Promise<void>;
}
export {};
