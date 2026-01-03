"use client";

import Link from "next/link";
import { useAgents } from "@/lib/hooks";

export default function AgentsPage() {
  const { agents, loading, error } = useAgents();

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
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-3xl font-bold">Agents</h1>
        </div>

        {error && (
          <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
            <p className="font-medium">Error loading agents</p>
            <p className="text-sm">{error.message}</p>
            <p className="text-xs mt-2 text-gray-600">
              Make sure the API server is running at{" "}
              {process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}
            </p>
          </div>
        )}

        {loading && !error ? (
          <div className="flex justify-center items-center h-40">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        ) : (
          <div className="grid gap-4">
            {agents.length === 0 ? (
              <div className="bg-white p-8 rounded-lg shadow-sm text-center text-gray-500">
                <p className="text-lg mb-2">No agents registered</p>
                <p className="text-sm">
                  Agents will appear here once they connect to the server
                </p>
              </div>
            ) : (
              agents.map((agent) => (
                <Link key={agent.id} href={`/agents/${agent.id}`}>
                  <div className="bg-white p-6 rounded-lg shadow-sm hover:shadow-md transition cursor-pointer">
                    <div className="flex items-start justify-between mb-3">
                      <div>
                        <h2 className="text-xl font-semibold">{agent.name}</h2>
                        <p className="text-sm text-gray-500">ID: {agent.id}</p>
                      </div>
                      <div
                        className={`px-3 py-1 rounded-full text-sm font-medium ${
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

                    {agent.capabilities && agent.capabilities.length > 0 && (
                      <div className="mb-3">
                        <p className="text-sm font-medium text-gray-700 mb-2">
                          Capabilities:
                        </p>
                        <div className="flex flex-wrap gap-2">
                          {agent.capabilities.map((cap) => (
                            <span
                              key={cap}
                              className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded"
                            >
                              {cap}
                            </span>
                          ))}
                        </div>
                      </div>
                    )}

                    <div className="text-xs text-gray-500 space-y-1">
                      {agent.registered_at && (
                        <p>
                          Registered:{" "}
                          {new Date(agent.registered_at).toLocaleString()}
                        </p>
                      )}
                      {agent.last_heartbeat && (
                        <p>
                          Last heartbeat:{" "}
                          {new Date(agent.last_heartbeat).toLocaleString()}
                        </p>
                      )}
                    </div>
                  </div>
                </Link>
              ))
            )}
          </div>
        )}
      </main>
    </div>
  );
}
