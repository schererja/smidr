export interface Config {
    project_name?: string;
    target_image?: string;
    machine?: string;
    build_directory?: string;
    download_directory?: string;
    [key: string]: any;
}
export declare class ConfigManager {
    loadConfig(configPath: string): Promise<Config>;
    saveConfig(configPath: string, config: Config): Promise<void>;
    validateConfig(config: Config): boolean;
}
