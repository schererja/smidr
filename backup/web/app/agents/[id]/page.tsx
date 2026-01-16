"use client";

import Link from "next/link";
import { useAgent } from "@/lib/hooks";
import { useParams } from "next/navigation";

export default function AgentDetailsPage() {
  const params = useParams();
  const agentID = params.id as string;
  const { agent, loading, error } = useAgent(agentID);

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b">
        <nav className="container mx-auto px-4 py-4 flex items-center justify-between">
          <Link href="/" className="text-2xl font-bold text-blue-600">
            Smidr
          </Link>
          <div className="flex items-center space-x-6">
            <Link
              href="/dashboard"
              className="text-gray-600 hover:text-gray-900"
            >
              Dashboard
            </Link>
            <Link href="/agents" className="text-blue-600 font-medium">
              Agents
            </Link>
          </div>
        </nav>
      </header>

      <main className="container mx-auto px-4 py-8">
        <div className="mb-6">
          <Link href="/agents" className="text-blue-600 hover:text-blue-700">
            ← Back to Agents
          </Link>
        </div>

        {error && (
          <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
            <p className="font-medium">Error loading agent</p>
            <p className="text-sm">{error.message}</p>
          </div>
        )}

        {loading ? (
          <div className="flex justify-center items-center h-40">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        ) : agent ? (
          <div className="bg-white rounded-lg shadow-sm p-8">
            <div className="flex justify-between items-start mb-6">
              <div>
                <h1 className="text-3xl font-bold mb-2">{agent.name}</h1>
                <p className="text-gray-600">ID: {agent.id}</p>
              </div>
              <div
                className={`px-4 py-2 rounded-full text-lg font-medium ${
                  agent.status === "healthy"
                    ? "bg-green-100 text-green-800"
                    : agent.status === "offline"
                    ? "bg-gray-100 text-gray-800"
                    : "bg-yellow-100 text-yellow-800"
                }`}
              >
                {agent.status || "unknown"}
              </div>
            </div>

            <div className="grid md:grid-cols-2 gap-8">
              <div>
                <h2 className="text-xl font-semibold mb-4">Capabilities</h2>
                {agent.capabilities && agent.capabilities.length > 0 ? (
                  <div className="space-y-2">
                    {agent.capabilities.map((cap) => (
                      <div
                        key={cap}
                        className="bg-blue-50 p-3 rounded border border-blue-200"
                      >
                        {cap}
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-gray-500">No capabilities listed</p>
                )}
              </div>

              <div>
                <h2 className="text-xl font-semibold mb-4">Information</h2>
                <div className="space-y-3 text-gray-700">
                  {agent.registered_at && (
                    <div>
                      <p className="text-sm font-medium text-gray-600">
                        Registered
                      </p>
                      <p>{new Date(agent.registered_at).toLocaleString()}</p>
                    </div>
                  )}
                  {agent.last_heartbeat && (
                    <div>
                      <p className="text-sm font-medium text-gray-600">
                        Last Heartbeat
                      </p>
                      <p>{new Date(agent.last_heartbeat).toLocaleString()}</p>
                    </div>
                  )}
                </div>
              </div>
            </div>

            {agent.metadata && Object.keys(agent.metadata).length > 0 && (
              <div className="mt-8">
                <h2 className="text-xl font-semibold mb-4">Metadata</h2>
                <div className="bg-gray-50 p-4 rounded border border-gray-200 font-mono text-sm">
                  <pre>{JSON.stringify(agent.metadata, null, 2)}</pre>
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="bg-white p-8 rounded-lg shadow-sm text-center text-gray-500">
            <p className="text-lg">Agent not found</p>
          </div>
        )}
      </main>
    </div>
  );
}
