using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace Smidr.Daemon.Core.Configuration
{
    public class DaemonConfiguration
    {
        public string Hostname { get; set; } = "localhost";
        public int Port { get; set; } = 50051;
        public string DatabasePath { get; set; } = "/var/lib/smidr/daemon.db";
        public string LogLevel { get; set; } = "Information";
        public int MaxConnections { get; set; } = 4;
        public string WorkspaceRoot { get; set; } = "/var/lib/smidr/workspaces";
        public DockerConfiguration Docker { get; set; } = new();

    }

}
