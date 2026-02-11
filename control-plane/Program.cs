using System.Security.Cryptography.X509Certificates;
using Microsoft.AspNetCore.Server.Kestrel.Core;
using Microsoft.AspNetCore.Server.Kestrel.Https;
using ControlPlane.Data;
using ControlPlane.Services;
using ControlPlane.Middleware;
using Microsoft.EntityFrameworkCore;

var builder = WebApplication.CreateBuilder(args);

builder.WebHost.ConfigureKestrel(options =>
{
    options.ListenAnyIP(5001, listenOptions =>
    {
        listenOptions.UseHttps(httpsOptions =>
        {
            httpsOptions.ClientCertificateMode = ClientCertificateMode.AllowCertificate;
            httpsOptions.CheckCertificateRevocation = false;
            httpsOptions.SslProtocols = System.Security.Authentication.SslProtocols.Tls12 | System.Security.Authentication.SslProtocols.Tls13;
            httpsOptions.ClientCertificateValidation = (cert, chain, errors) => true;
        });
    });
});

var dataDir = Path.Combine(builder.Environment.ContentRootPath, "data");
Directory.CreateDirectory(dataDir);

var usePostgres = builder.Configuration.GetValue<bool>("UsePostgres");
var connectionString = usePostgres
    ? builder.Configuration.GetConnectionString("PostgreSQL") ?? "Host=localhost;Database=smidr;Username=smidr;Password=smidr"
    : $"Data Source={Path.Combine(dataDir, "controlplane.db")}";

builder.Services.AddDbContext<ControlPlaneDbContext>(options =>
{
    if (usePostgres)
        options.UseNpgsql(connectionString);
    else
        options.UseSqlite(connectionString);
});

var caService = new CaService(dataDir);
builder.Services.AddSingleton(caService);
builder.Services.AddSingleton<MtlsValidationService>();
builder.Services.AddScoped<HealthEvaluationService>();

builder.Services.AddCors(options =>
{
    options.AddDefaultPolicy(policy =>
    {
        policy.WithOrigins("http://localhost:5173", "http://localhost:3000", "http://localhost:3001", "https://localhost:5173")
              .AllowCredentials()
              .AllowAnyMethod()
              .AllowAnyHeader();
    });
});

builder.Services.AddControllers();
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

var app = builder.Build();

using (var scope = app.Services.CreateScope())
{
    var db = scope.ServiceProvider.GetRequiredService<ControlPlaneDbContext>();
    db.Database.EnsureCreated();
}

app.UseSwagger();
app.UseSwaggerUI();

app.UseCors();

app.UseMiddleware<MtlsAuthenticationMiddleware>();

app.MapControllers();

app.Run();
