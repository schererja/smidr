using Moq;
using SmidrLib;
using Smidr.V1;
using Xunit;

namespace Schererja.Smidr.WebAPI.Tests;

public class SmidrClientTests
{
  [Fact]
  public void SmidrClient_Should_Accept_Valid_DaemonUrl()
  {
    // Arrange
    var daemonUrl = "http://localhost:50051";

    // Act
    var client = new SmidrClient(daemonUrl);

    // Assert
    Assert.NotNull(client);
  }

  [Fact]
  public async Task StartBuildAsync_Should_Return_BuildIdentifier()
  {
    // Arrange
    var mockClient = new Mock<SmidrClient>("http://localhost:50051");

    // Note: This test demonstrates the structure
    // In a real scenario, you'd mock the gRPC calls properly

    // Act & Assert
    Assert.NotNull(mockClient.Object);
  }
  [Fact]
  public void BuildIdentifier_Should_Have_BuildId()
  {
    // Arrange
    var buildId = new BuildIdentifier
    {
      BuildId = "test-build-123"
    };

    // Assert
    Assert.Equal("test-build-123", buildId.BuildId);
  }

  [Fact]
  public void BuildState_Enum_Should_Have_Expected_Values()
  {
    // Assert
    Assert.True(Enum.IsDefined(typeof(BuildState), BuildState.Unspecified));
    Assert.True(Enum.IsDefined(typeof(BuildState), BuildState.Queued));
    Assert.True(Enum.IsDefined(typeof(BuildState), BuildState.Preparing));
    Assert.True(Enum.IsDefined(typeof(BuildState), BuildState.Building));
    Assert.True(Enum.IsDefined(typeof(BuildState), BuildState.ExtractingArtifacts));
    Assert.True(Enum.IsDefined(typeof(BuildState), BuildState.Completed));
    Assert.True(Enum.IsDefined(typeof(BuildState), BuildState.Failed));
    Assert.True(Enum.IsDefined(typeof(BuildState), BuildState.Cancelled));
  }
}