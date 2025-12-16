using Serilog;
using Smidr.Daemon.Commands;

var logger = new LoggerConfiguration()
    .WriteTo.Console()
    .WriteTo.File("logs/daemon.log", rollingInterval: RollingInterval.Day)
    .CreateLogger();

try
{
    var commandArgs = args.Length == 0 ? new[] { "--help" } : args;
    var command = commandArgs[0].ToLower();

    switch (command)
    {
        case "init":
            // await InitCommand.ExecuteAsync(commandArgs.Skip(1).ToArray(), logger);
            break;
        case "build":
            await BuildCommand.ExecuteAsync(commandArgs.Skip(1).ToArray(), logger);
            break;
        case "--help":
        case "-h":
        case "help":
            PrintHelp();
            break;
        default:
            logger.Error("Unknown command: {Command}", command);
            PrintHelp();
            Environment.Exit(1);
            break;
    }
}
catch (Exception ex)
{
    logger.Fatal(ex, "Fatal error");
    Environment.Exit(1);
}

void PrintHelp()
{
    Console.WriteLine("""
        Smidr Daemon v0.1.0

        Usage: smidr <command> [options]

        Commands:
          init       Initialize a new smidr project
          build      Start a build
          help       Show this help message

        Examples:
          smidr init -o smidr.yaml
          smidr build --config smidr.yaml
        """);
}
