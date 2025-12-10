import fs from "fs/promises";
import yaml from "yaml";
export class ConfigManager {
    async loadConfig(configPath) {
        try {
            const content = await fs.readFile(configPath, "utf-8");
            const config = yaml.parse(content);
            return config;
        }
        catch (error) {
            throw new Error(`Failed to load config from ${configPath}: ${error.message}`);
        }
    }
    async saveConfig(configPath, config) {
        try {
            const content = yaml.stringify(config);
            await fs.writeFile(configPath, content, "utf-8");
        }
        catch (error) {
            throw new Error(`Failed to save config to ${configPath}: ${error.message}`);
        }
    }
    validateConfig(config) {
        return !!(config.project_name && config.target_image && config.machine);
    }
}
