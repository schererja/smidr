using Microsoft.EntityFrameworkCore;
using Serilog;
using Smidr.ControlPlane.Api.GraphQL;
using Smidr.ControlPlane.Api.Middleware;
using Smidr.ControlPlane.Api.Services;
using Smidr.ControlPlane.Application.Common;
using Smidr.ControlPlane.Infrastructure.Persistence;

var builder = WebApplication.CreateBuilder(args);

// Serilog
Log.Logger = new LoggerConfiguration()
    .ReadFrom.Configuration(builder.Configuration)
    .Enrich.FromLogContext()
    .WriteTo.Console()
    .CreateLogger();

builder.Host.UseSerilog();

// Database
builder.Services.AddDbContext<SmidrDbContext>(options =>
{
    var connectionString = builder.Configuration.GetConnectionString("DefaultConnection")
        ?? "Data Source=smidr.db";
    options.UseSqlite(connectionString);
});

// Multi-tenancy
builder.Services.AddScoped<ITenantContext, TenantContext>();

// gRPC
builder.Services.AddGrpc(options =>
{
    options.EnableDetailedErrors = builder.Environment.IsDevelopment();
});
builder.Services.AddGrpcReflection();

// GraphQL
builder.Services
    .AddGraphQLServer()
    .AddQueryType<Query>()
    .AddMutationType<Mutation>()
    .AddFiltering()
    .AddSorting()
    .RegisterDbContext<SmidrDbContext>(DbContextKind.Pooled);

// CORS for GraphQL
builder.Services.AddCors(options =>
{
    options.AddPolicy("AllowGraphQL", policy =>
    {
        policy.AllowAnyOrigin()
              .AllowAnyMethod()
              .AllowAnyHeader();
    });
});

var app = builder.Build();

// Ensure database is created
using (var scope = app.Services.CreateScope())
{
    var dbContext = scope.ServiceProvider.GetRequiredService<SmidrDbContext>();
    dbContext.Database.EnsureCreated();
}

// Middleware
app.UseSerilogRequestLogging();
app.UseCors("AllowGraphQL");
app.UseMiddleware<TenantResolutionMiddleware>();

// gRPC endpoints
app.MapGrpcService<AgentGrpcService>();

// gRPC reflection (for development)
if (app.Environment.IsDevelopment())
{
    app.MapGrpcReflectionService();
}

// GraphQL endpoint
app.MapGraphQL("/graphql");

// Health check
app.MapGet("/health", () => Results.Ok(new { status = "healthy", timestamp = DateTime.UtcNow }));

// TLS info
app.MapGet("/", () => Results.Ok(new
{
    service = "Smidr Control Plane",
    version = "1.0.0",
    endpoints = new
    {
        grpc = "Port 5000 (agents)",
        graphql = "/graphql (UI/API)",
        health = "/health"
    },
    features = new[] { "Multi-tenancy", "TLS", "gRPC", "GraphQL" }
}));

app.Run();
