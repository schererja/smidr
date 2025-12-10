import { SmidrServer } from "./server.js";
import { Logger } from "./logger.js";
import { ConfigManager } from "./config.js";
import { DB } from "./database.js";
import { DockerManager } from "./docker.js";
const logger = new Logger("smidr-daemon");
const configManager = new ConfigManager();
const database = new DB(process.env.DB_PATH || "smidr.db");
const dockerManager = new DockerManager(logger);
const server = new SmidrServer(process.env.HOST || "0.0.0.0", parseInt(process.env.PORT || "50051"), logger, configManager, database, dockerManager);
async function main() {
    try {
        // Initialize database
        await database.initialize();
        logger.info("Database initialized");
        // Check Docker availability
        const dockerRunning = await dockerManager.isRunning();
        if (!dockerRunning) {
            logger.warn("Docker daemon is not available. Builds will fail.");
        }
        await server.start();
        logger.info("Smidr daemon started successfully");
        // Graceful shutdown
        process.on("SIGTERM", async () => {
            logger.info("Received SIGTERM, shutting down gracefully...");
            await server.stop();
            await database.close();
            process.exit(0);
        });
        process.on("SIGINT", async () => {
            logger.info("Received SIGINT, shutting down gracefully...");
            await server.stop();
            await database.close();
            process.exit(0);
        });
    }
    catch (error) {
        logger.error("Failed to start daemon", error);
        process.exit(1);
    }
}
main();
