using System;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace controlplane.Migrations
{
    /// <inheritdoc />
    public partial class InitialCreate : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateTable(
                name: "AgentBaselines",
                columns: table => new
                {
                    Id = table.Column<int>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    AgentId = table.Column<string>(type: "TEXT", nullable: false),
                    MetricName = table.Column<string>(type: "TEXT", nullable: false),
                    Mean = table.Column<double>(type: "REAL", nullable: false),
                    StdDev = table.Column<double>(type: "REAL", nullable: false),
                    Min = table.Column<double>(type: "REAL", nullable: false),
                    Max = table.Column<double>(type: "REAL", nullable: false),
                    SampleCount = table.Column<int>(type: "INTEGER", nullable: false),
                    CreatedAt = table.Column<DateTime>(type: "TEXT", nullable: false),
                    UpdatedAt = table.Column<DateTime>(type: "TEXT", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_AgentBaselines", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "Agents",
                columns: table => new
                {
                    Id = table.Column<string>(type: "TEXT", nullable: false),
                    Hostname = table.Column<string>(type: "TEXT", nullable: false),
                    Token = table.Column<string>(type: "TEXT", nullable: true),
                    CertificatePem = table.Column<string>(type: "TEXT", nullable: true),
                    RevokedAt = table.Column<DateTime>(type: "TEXT", nullable: true),
                    RegisteredAt = table.Column<DateTime>(type: "TEXT", nullable: false),
                    LastHeartbeatAt = table.Column<DateTime>(type: "TEXT", nullable: true),
                    CurrentHealth = table.Column<string>(type: "TEXT", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Agents", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "HealthStates",
                columns: table => new
                {
                    Id = table.Column<int>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    AgentId = table.Column<string>(type: "TEXT", nullable: false),
                    Status = table.Column<string>(type: "TEXT", nullable: false),
                    Reason = table.Column<string>(type: "TEXT", nullable: true),
                    ChangedAt = table.Column<DateTime>(type: "TEXT", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_HealthStates", x => x.Id);
                });

            migrationBuilder.CreateTable(
                name: "Heartbeats",
                columns: table => new
                {
                    Id = table.Column<int>(type: "INTEGER", nullable: false)
                        .Annotation("Sqlite:Autoincrement", true),
                    AgentId = table.Column<string>(type: "TEXT", nullable: false),
                    Timestamp = table.Column<DateTime>(type: "TEXT", nullable: false),
                    UptimeSeconds = table.Column<double>(type: "REAL", nullable: false),
                    LoadAverage1m = table.Column<double>(type: "REAL", nullable: false),
                    MemoryUsedPct = table.Column<double>(type: "REAL", nullable: false),
                    DiskUsedPct = table.Column<double>(type: "REAL", nullable: false),
                    ProcessCount = table.Column<int>(type: "INTEGER", nullable: false),
                    ReceivedAt = table.Column<DateTime>(type: "TEXT", nullable: false)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_Heartbeats", x => x.Id);
                });

            migrationBuilder.CreateIndex(
                name: "IX_AgentBaselines_AgentId_MetricName",
                table: "AgentBaselines",
                columns: new[] { "AgentId", "MetricName" },
                unique: true);

            migrationBuilder.CreateIndex(
                name: "IX_HealthStates_AgentId",
                table: "HealthStates",
                column: "AgentId");

            migrationBuilder.CreateIndex(
                name: "IX_HealthStates_ChangedAt",
                table: "HealthStates",
                column: "ChangedAt");

            migrationBuilder.CreateIndex(
                name: "IX_Heartbeats_AgentId",
                table: "Heartbeats",
                column: "AgentId");

            migrationBuilder.CreateIndex(
                name: "IX_Heartbeats_Timestamp",
                table: "Heartbeats",
                column: "Timestamp");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "AgentBaselines");

            migrationBuilder.DropTable(
                name: "Agents");

            migrationBuilder.DropTable(
                name: "HealthStates");

            migrationBuilder.DropTable(
                name: "Heartbeats");
        }
    }
}
