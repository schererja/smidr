import { createPromiseClient, PromiseClient } from "@connectrpc/connect";
import { createGrpcTransport } from "@connectrpc/connect-node";
import { create } from "@bufbuild/protobuf";
import { BuildService } from "./Generated/builds_connect.js";
import { LogService } from "./Generated/logs_connect.js";
import { ArtifactService } from "./Generated/artifacts_connect.js";
import type {
  StartBuildRequest,
  BuildStatusResponse,
  ListBuildsRequest,
  ListBuildsResponse,
  CancelBuildResponse,
} from "./Generated/builds_pb.js";
import type { ListArtifactsResponse } from "./Generated/artifacts_pb.js";
import type { LogEntry, StreamBuildLogsRequest } from "./Generated/logs_pb.js";
import { BuildIdentifierSchema, BuildState } from "./Generated/common_pb.js";
import type { BuildIdentifier } from "./Generated/common_pb.js";

export interface SmidrClientOptions {
  /**
   * The address of the Smidr daemon (e.g., "http://localhost:50051")
   */
  address: string;

  /**
   * Optional HTTP/2 transport options
   */
  httpVersion?: "1.1" | "2";
}

/**
 * Client for interacting with the Smidr gRPC daemon using Connect-RPC.
 */
export class SmidrClient {
  private buildClient: PromiseClient<typeof BuildService>;
  private logClient: PromiseClient<typeof LogService>;
  private artifactClient: PromiseClient<typeof ArtifactService>;

  constructor(options: SmidrClientOptions) {
    const transport = createGrpcTransport({
      baseUrl: options.address,
      httpVersion: options.httpVersion ?? "2",
    });

    this.buildClient = createPromiseClient(BuildService, transport);
    this.logClient = createPromiseClient(LogService, transport);
    this.artifactClient = createPromiseClient(ArtifactService, transport);
  }

  /**
   * Starts a new build with the specified configuration.
   */
  async startBuild(
    request: Partial<StartBuildRequest>
  ): Promise<BuildStatusResponse> {
    return this.buildClient.startBuild(request);
  }

  /**
   * Gets the current status of a build.
   */
  async getBuildStatus(buildId: string): Promise<BuildStatusResponse> {
    return this.buildClient.getBuildStatus({
      buildIdentifier: create(BuildIdentifierSchema, { buildId }),
    });
  }

  /**
   * Lists all builds matching the specified filters.
   */
  async listBuilds(
    request?: Partial<ListBuildsRequest>
  ): Promise<ListBuildsResponse> {
    return this.buildClient.listBuilds(request ?? {});
  }

  /**
   * Cancels a running build.
   */
  async cancelBuild(buildId: string): Promise<CancelBuildResponse> {
    return this.buildClient.cancelBuild({
      buildIdentifier: create(BuildIdentifierSchema, { buildId }),
    });
  }

  /**
   * Lists artifacts for a completed build.
   */
  async listArtifacts(buildId: string): Promise<ListArtifactsResponse> {
    return this.artifactClient.listArtifacts({
      buildIdentifier: create(BuildIdentifierSchema, { buildId }),
    });
  }

  /**
   * Streams logs from a build in real-time.
   */
  async *streamLogs(
    buildId: string,
    follow: boolean = false
  ): AsyncIterable<LogEntry> {
    const request: StreamBuildLogsRequest = {
      buildIdentifier: create(BuildIdentifierSchema, { buildId }),
      follow,
    };

    for await (const logEntry of this.logClient.streamBuildLogs(request)) {
      yield logEntry;
    }
  }

  /**
   * Direct access to the BuildService client for advanced operations.
   */
  get builds(): PromiseClient<typeof BuildService> {
    return this.buildClient;
  }

  /**
   * Direct access to the LogService client for advanced operations.
   */
  get logs(): PromiseClient<typeof LogService> {
    return this.logClient;
  }

  /**
   * Direct access to the ArtifactService client for advanced operations.
   */
  get artifacts(): PromiseClient<typeof ArtifactService> {
    return this.artifactClient;
  }
}

// Re-export commonly used types
export { BuildState, BuildIdentifier };
export type {
  BuildStatusResponse,
  StartBuildRequest,
  ListBuildsResponse,
  LogEntry,
  ListArtifactsResponse,
};
