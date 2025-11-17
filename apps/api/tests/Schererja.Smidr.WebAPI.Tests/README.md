# Schererja.Smidr.WebAPI.Tests

Unit and integration tests for the Smidr Web API.

## Overview

This test project uses xUnit for testing the REST API gateway that communicates with the Smidr gRPC daemon.

## Test Categories

### Integration Tests (`WebAPIIntegrationTests.cs`)

Tests that verify the API endpoints work correctly:

- Swagger documentation is available
- API endpoints exist and respond appropriately
- Endpoints handle missing daemon gracefully

### Unit Tests (`SmidrClientTests.cs`)

Tests for the SmidrLib client and protobuf models:

- Client initialization
- BuildState enum validation
- Message structure validation

## Running Tests

From the `apps/api` directory:

```bash
# Run all tests
dotnet test

# Run with verbose output
dotnet test -v detailed

# Run specific test
dotnet test --filter "FullyQualifiedName~WebAPIIntegrationTests"

# Run with coverage (requires coverlet.collector)
dotnet test --collect:"XPlat Code Coverage"
```

## Test Dependencies

- **xUnit**: Test framework
- **Microsoft.AspNetCore.Mvc.Testing**: For integration testing ASP.NET Core apps
- **Moq**: Mocking framework for unit tests

## Notes

- Integration tests require the daemon to be running for full functionality tests
- Tests that hit endpoints without a running daemon expect service unavailable responses
- The `Program` class is made public via `public partial class Program { }` to enable WebApplicationFactory testing
