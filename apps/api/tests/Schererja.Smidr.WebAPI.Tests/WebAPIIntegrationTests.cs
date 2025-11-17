using Microsoft.AspNetCore.Mvc.Testing;
using System.Net;
using System.Net.Http.Json;
using Xunit;

namespace Schererja.Smidr.WebAPI.Tests;

public class WebAPIIntegrationTests : IClassFixture<WebApplicationFactory<Program>>
{
  private readonly HttpClient _client;

  public WebAPIIntegrationTests(WebApplicationFactory<Program> factory)
  {
    _client = factory.CreateClient();
  }

  [Fact]
  public async Task Swagger_Endpoint_Returns_Success()
  {
    // Act
    var response = await _client.GetAsync("/swagger/v1/swagger.json");

    // Assert
    response.EnsureSuccessStatusCode();
    Assert.Equal("application/json", response.Content.Headers.ContentType?.MediaType);
  }

  [Fact]
  public async Task Root_Endpoint_Returns_Swagger_UI()
  {
    // Act
    var response = await _client.GetAsync("/");

    // Assert
    response.EnsureSuccessStatusCode();
    var content = await response.Content.ReadAsStringAsync();
    Assert.Contains("Swagger UI", content);
  }

  [Fact]
  public async Task StartBuild_Endpoint_Exists()
  {
    // Arrange
    var buildRequest = new
    {
      ConfigPath = "/path/to/config.yaml",
      Target = "core-image-minimal"
    };

    // Act - We expect it to fail because daemon isn't running
    // but we're testing the endpoint exists and validates input
    var response = await _client.PostAsJsonAsync("/api/builds", buildRequest);

    // Assert - Should be Service Unavailable (503) or similar, not NotFound
    Assert.NotEqual(HttpStatusCode.NotFound, response.StatusCode);
  }

  [Fact]
  public async Task GetBuildStatus_Endpoint_Exists()
  {
    // Act
    var response = await _client.GetAsync("/api/builds/test-build-id");

    // Assert - Should not be NotFound route, but may be NotFound build
    Assert.NotEqual(HttpStatusCode.NotFound, response.StatusCode);
  }

  [Fact]
  public async Task ListBuilds_Endpoint_Returns_Success_Or_ServiceUnavailable()
  {
    // Act
    var response = await _client.GetAsync("/api/builds");

    // Assert
    Assert.True(
        response.StatusCode == HttpStatusCode.OK ||
        response.StatusCode == HttpStatusCode.ServiceUnavailable ||
        response.IsSuccessStatusCode,
        $"Expected success or service unavailable, but got {response.StatusCode}"
    );
  }
}
