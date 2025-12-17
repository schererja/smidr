using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.Extensions.Configuration;
using YamlDotNet.Serialization.NamingConventions;

namespace Smidr.Daemon.Core.Configuration
{
    /// <summary>
    /// Loads the DaemonConfiguration from various sources with the following priority:
    /// 1. Environment Variables
    /// 2. YAML Configuration File
    /// 3. Default Values
    /// </summary>
    public class ConfigurationLoader
    {
        public static DaemonConfiguration LoadConfiguration(IConfiguration configuration)
        {
            var config = new DaemonConfiguration();

            // Load configuration values from the provided IConfiguration - highest priority
            config.Hostname = configuration["SMIDR_DAEMON_HOSTNAME"] ?? config.Hostname;
            config.Port = int.TryParse(configuration["SMIDR_DAEMON_PORT"] ?? "", out var port) ? port : config.Port;
            config.DatabasePath = configuration["SMIDR_DAEMON_DATABASE_PATH"] ?? config.DatabasePath;
            config.LogLevel = configuration["SMIDR_DAEMON_LOG_LEVEL"] ?? config.LogLevel;
            config.MaxConnections = int.TryParse(configuration["SMIDR_DAEMON_MAX_CONNECTIONS"] ?? "", out var maxConns) ? maxConns : config.MaxConnections;
            config.WorkspaceRoot = configuration["SMIDR_DAEMON_WORKSPACE_ROOT"] ?? config.WorkspaceRoot;
            // Docker configuration
            config.Docker.DefaultImage = configuration["SMIDR_DAEMON_DOCKER_DEFAULT_IMAGE"] ?? config.Docker.DefaultImage;
            config.Docker.SocketPath = configuration["SMIDR_DAEMON_DOCKER_SOCKET_PATH"] ?? config.Docker.SocketPath;
            config.Docker.MemoryLimitMb = long.TryParse(configuration["SMIDR_DAEMON_DOCKER_MEMORY_LIMIT_MB"] ?? "", out var memLimit) ? memLimit : config.Docker.MemoryLimitMb;
            config.Docker.CpuLimit = int.TryParse(configuration["SMIDR_DAEMON_DOCKER_CPU_LIMIT"] ?? "", out var cpuLimit) ? cpuLimit : config.Docker.CpuLimit;

            // Load from YAML file if exists
            var yamlFilePath = configuration["SMIDR_DAEMON_CONFIG_FILE"] ?? "/etc/smidr/config.yaml";
            if (System.IO.File.Exists(yamlFilePath))
            {
                LoadFromYamlFile(yamlFilePath, config);
            }
            return config;
        }

        private static void LoadFromYamlFile(string yamlFilePath, DaemonConfiguration config)
        {
            var yaml = File.ReadAllText(yamlFilePath);
            var deserializer = new YamlDotNet.Serialization.DeserializerBuilder().WithNamingConvention(UnderscoredNamingConvention.Instance).Build();
            var yamlConfig = deserializer.Deserialize<DaemonConfiguration>(yaml);

            // Merge yaml config into existing config, only overwriting defaults
            if (yamlConfig != null)
            {
                config.Hostname = yamlConfig.Hostname ?? config.Hostname;
                config.Port = yamlConfig.Port != 0 ? yamlConfig.Port : config.Port;
                config.DatabasePath = yamlConfig.DatabasePath ?? config.DatabasePath;
                config.LogLevel = yamlConfig.LogLevel ?? config.LogLevel;
                config.MaxConnections = yamlConfig.MaxConnections != 0 ? yamlConfig.MaxConnections : config.MaxConnections;
                config.WorkspaceRoot = yamlConfig.WorkspaceRoot ?? config.WorkspaceRoot;

                if (yamlConfig.Docker != null)
                {
                    config.Docker.DefaultImage = yamlConfig.Docker.DefaultImage ?? config.Docker.DefaultImage;
                    config.Docker.SocketPath = yamlConfig.Docker.SocketPath ?? config.Docker.SocketPath;
                    config.Docker.MemoryLimitMb = yamlConfig.Docker.MemoryLimitMb != 0 ? yamlConfig.Docker.MemoryLimitMb : config.Docker.MemoryLimitMb;
                    config.Docker.CpuLimit = yamlConfig.Docker.CpuLimit != 0 ? yamlConfig.Docker.CpuLimit : config.Docker.CpuLimit;
                }
            }
        }
    }
}
