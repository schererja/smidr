using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace Smidr.Daemon.Core.Configuration
{
    public class SecurityConfiguration
    {
        public bool EnableTls { get; set; } = false;
        public string? CertificatePath { get; set; } = null;
        public string? PrivateKeyPath { get; set; } = null;
        public bool RequireClientCertificates { get; set; } = false;
    }
}
