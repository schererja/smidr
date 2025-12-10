export declare class Logger {
    private logger;
    constructor(label?: string);
    info(message: string, meta?: any): void;
    warn(message: string, meta?: any): void;
    error(message: string, error?: any): void;
    debug(message: string, meta?: any): void;
}
