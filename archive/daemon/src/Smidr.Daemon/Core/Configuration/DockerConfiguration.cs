using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace Smidr.Daemon.Core.Configuration
{
    public class DockerConfiguration
    {
        public string DefaultImage { get; set; } = "crops/yocto:ubuntu-22.04-builder";
        public string SocketPath { get; set; } = "/var/run/docker.sock";
        public long MemoryLimitMb { get; set; } = 8589934592; // 8 GB
        public int CpuLimit { get; set; } = 2;
    }
}
