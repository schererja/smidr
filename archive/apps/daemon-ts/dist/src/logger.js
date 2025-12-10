import winston from "winston";
export class Logger {
    constructor(label = "smidr") {
        this.logger = winston.createLogger({
            level: process.env.LOG_LEVEL || "info",
            format: winston.format.combine(winston.format.timestamp(), winston.format.errors({ stack: true }), winston.format.json()),
            defaultMeta: { service: label },
            transports: [
                new winston.transports.Console({
                    format: winston.format.combine(winston.format.colorize(), winston.format.simple()),
                }),
                new winston.transports.File({ filename: "error.log", level: "error" }),
                new winston.transports.File({ filename: "combined.log" }),
            ],
        });
    }
    info(message, meta) {
        this.logger.info(message, meta);
    }
    warn(message, meta) {
        this.logger.warn(message, meta);
    }
    error(message, error) {
        this.logger.error(message, { error });
    }
    debug(message, meta) {
        this.logger.debug(message, meta);
    }
}
