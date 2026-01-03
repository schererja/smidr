"use client";

import Link from "next/link";
import { useAgents } from "@/lib/hooks";

export default function Dashboard() {
  const { agents, loading, error } = useAgents();

  const activeAgents = agents.filter((a) => a.status === "healthy").length;
  const totalAgents = agents.length;

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b">
        <nav className="container mx-auto px-4 py-4 flex items-center justify-between">
          <Link href="/" className="text-2xl font-bold text-blue-600">
            Smidr
          </Link>
          <div className="flex items-center space-x-6">
            <Link href="/dashboard" className="text-blue-600 font-medium">
              Dashboard
            </Link>
            <Link href="/agents" className="text-gray-600 hover:text-gray-900">
              Agents
            </Link>
            <Link
              href="/settings"
              className="text-gray-600 hover:text-gray-900"
            >
              Settings
            </Link>
          </div>
        </nav>
      </header>

      <main className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-8">Dashboard</h1>

        {error && (
          <div className="mb-4 p-4 bg-yellow-50 border border-yellow-200 rounded-lg text-yellow-700">
            <p className="font-medium">Warning: Unable to connect to API</p>
            <p className="text-sm">{error.message}</p>
            <p className="text-xs mt-2">
              Make sure the backend server is running at{" "}
              {process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}
            </p>
          </div>
        )}

        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <Link href="/agents">
            <div className="bg-white p-6 rounded-lg shadow-sm hover:shadow-md transition cursor-pointer">
              <div className="text-sm text-gray-600 mb-1">Total Agents</div>
              <div className="text-3xl font-bold">
                {loading ? "-" : totalAgents}
              </div>
            </div>
          </Link>
          <div className="bg-white p-6 rounded-lg shadow-sm">
            <div className="text-sm text-gray-600 mb-1">Active Agents</div>
            <div className="text-3xl font-bold text-green-600">
              {loading ? "-" : activeAgents}
            </div>
          </div>
          <div className="bg-white p-6 rounded-lg shadow-sm">
            <div className="text-sm text-gray-600 mb-1">Offline Agents</div>
            <div className="text-3xl font-bold text-gray-600">
              {loading ? "-" : totalAgents - activeAgents}
            </div>
          </div>
          <div className="bg-white p-6 rounded-lg shadow-sm">
            <div className="text-sm text-gray-600 mb-1">API Status</div>
            <div
              className={`text-3xl font-bold ${
                error ? "text-red-600" : "text-green-600"
              }`}
            >
              {loading ? "Loading..." : error ? "Error" : "Healthy"}
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm">
          <div className="p-6 border-b">
            <h2 className="text-xl font-semibold">Connected Agents</h2>
          </div>
          <div className="p-6">
            {loading && !error ? (
              <div className="flex justify-center py-8">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              </div>
            ) : agents.length === 0 ? (
              <div className="text-center text-gray-500 py-8">
                <p className="mb-2">No agents connected</p>
                <p className="text-sm">
                  Agents will appear here once they register with the server
                </p>
              </div>
            ) : (
              <div className="space-y-4">
                {agents.slice(0, 5).map((agent) => (
                  <Link key={agent.id} href={`/agents/${agent.id}`}>
                    <div className="flex items-center justify-between py-3 border-b last:border-b-0 hover:bg-gray-50 px-2 -mx-2 rounded cursor-pointer">
                      <div className="flex items-center space-x-4 flex-1">
                        <div
                          className={`w-3 h-3 rounded-full ${
                            agent.status === "healthy"
                              ? "bg-green-500"
                              : agent.status === "offline"
                              ? "bg-gray-400"
                              : "bg-yellow-500"
                          }`}
                        ></div>
                        <div>
                          <div className="font-medium">{agent.name}</div>
                          <div className="text-sm text-gray-500">
                            {agent.id}
                          </div>
                        </div>
                      </div>
                      <div className="text-sm text-gray-600 font-mono">
                        {agent.status || "unknown"}
                      </div>
                    </div>
                  </Link>
                ))}
                {agents.length > 5 && (
                  <div className="pt-4 border-t">
                    <Link
                      href="/agents"
                      className="text-blue-600 hover:text-blue-700 font-medium"
                    >
                      View all {agents.length} agents →
                    </Link>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
