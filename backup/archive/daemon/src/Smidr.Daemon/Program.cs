using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Serilog;
using Smidr.Daemon;
using Smidr.Daemon.Core.Configuration;
using Smidr.Daemon.Core.Scheduling;

// ...existing code...
// Replace everything with the Generic Host bootstrap:

Log.Logger = new LoggerConfiguration()
    .Enrich.FromLogContext()
    .WriteTo.Console()
    .WriteTo.File("logs/daemon.log", rollingInterval: RollingInterval.Day)
    .CreateLogger();

try
{
    var builder = Host.CreateDefaultBuilder(args)
        .UseSerilog()
        .ConfigureAppConfiguration((context, config) =>
        {
            config.AddEnvironmentVariables(prefix: "SMIDR_");
            // Optional: add JSON/YAML later
        })
        .ConfigureServices((context, services) =>
        {
            // Configuration loader
            services.AddSingleton<DaemonConfiguration>(sp =>
            {
                var cfg = sp.GetRequiredService<IConfiguration>();
                return ConfigurationLoader.LoadConfiguration(cfg);
            });

            // Core scheduling
            services.AddSingleton<BuildQueue>(sp =>
            {
                var dc = sp.GetRequiredService<DaemonConfiguration>();
                return new BuildQueue(dc.MaxConcurrentBuilds);
            });
            services.AddSingleton<BuildScheduler>();

            // Hosted daemon
            services.AddHostedService<DaemonHost>();
        });

    await builder.Build().RunAsync();
}
catch (Exception ex)
{
    Log.Fatal(ex, "Smidr daemon terminated unexpectedly");
}
finally
{
    await Log.CloseAndFlushAsync();
}
