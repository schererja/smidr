using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace controlplane.Migrations
{
    /// <inheritdoc />
    public partial class AddOSToAgent : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.AddColumn<string>(
                name: "OS",
                table: "Agents",
                type: "TEXT",
                nullable: true);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "OS",
                table: "Agents");
        }
    }
}
