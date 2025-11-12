// See https://aka.ms/new-console-template for more information


using System;
using Grpc.Net.Client;
using SmidrLib;



Console.WriteLine("Hello, Smidr!");

// Create a gRPC channel to the Smidr daemon
using var client = new SmidrClient("http://smidr-server.ik8labs.local:50051");
// Start a new build
var status = await client.StartBuildAsync(
    configPath: "smidr.yaml",
    target: "core-image-minimal",
    customer: "acme"
);
Console.WriteLine($"Build Status: {status}");
