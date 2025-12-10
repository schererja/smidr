"use client";

import { useEffect, useState } from "react";
import Link from "next/link";

interface Build {
  buildId: string;
  state: string;
  startTime: string;
  target: string;
}

export default function BuildsPage() {
  const [builds, setBuilds] = useState<Build[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchBuilds = async () => {
      try {
        const apiUrl =
          process.env.NEXT_PUBLIC_API_URL || "http://localhost:5285";
        const response = await fetch(`${apiUrl}/api/builds`);

        if (!response.ok) {
          throw new Error(`Failed to fetch builds: ${response.statusText}`);
        }

        const data = await response.json();
        setBuilds(data.builds || []);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to fetch builds");
      } finally {
        setLoading(false);
      }
    };

    fetchBuilds();

    // Refresh every 5 seconds
    const interval = setInterval(fetchBuilds, 5000);
    return () => clearInterval(interval);
  }, []);

  const getStateColor = (state: string) => {
    switch (state.toLowerCase()) {
      case "completed":
        return "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200";
      case "failed":
        return "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200";
      case "building":
      case "preparing":
        return "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200";
      case "queued":
        return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200";
      case "cancelled":
        return "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200";
      default:
        return "bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200";
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800">
      <div className="container mx-auto px-4 py-8">
        <div className="mb-8 flex items-center justify-between">
          <div>
            <Link
              href="/"
              className="mb-2 inline-block text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400"
            >
              ← Back to Home
            </Link>
            <h1 className="text-4xl font-bold text-gray-900 dark:text-white">
              Builds
            </h1>
          </div>
          <Link
            href="/builds/new"
            className="rounded-lg bg-blue-600 px-4 py-2 text-white transition-colors hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600"
          >
            + New Build
          </Link>
        </div>

        {loading && (
          <div className="rounded-lg bg-white p-8 text-center shadow-lg dark:bg-gray-800">
            <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-blue-600 border-r-transparent"></div>
            <p className="mt-4 text-gray-600 dark:text-gray-400">
              Loading builds...
            </p>
          </div>
        )}

        {error && (
          <div className="rounded-lg bg-red-50 p-4 text-red-800 dark:bg-red-900/20 dark:text-red-200">
            <p className="font-semibold">Error loading builds</p>
            <p className="text-sm">{error}</p>
          </div>
        )}

        {!loading && !error && builds.length === 0 && (
          <div className="rounded-lg bg-white p-8 text-center shadow-lg dark:bg-gray-800">
            <p className="text-gray-600 dark:text-gray-400">No builds found</p>
            <Link
              href="/builds/new"
              className="mt-4 inline-block text-blue-600 hover:text-blue-700 dark:text-blue-400"
            >
              Start your first build →
            </Link>
          </div>
        )}

        {!loading && !error && builds.length > 0 && (
          <div className="space-y-4">
            {builds.map((build) => (
              <Link
                key={build.buildId}
                href={`/builds/${build.buildId}`}
                className="block rounded-lg bg-white p-6 shadow-lg transition-all hover:shadow-xl dark:bg-gray-800"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="mb-2 flex items-center gap-3">
                      <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                        {build.buildId}
                      </h3>
                      <span
                        className={`rounded-full px-3 py-1 text-xs font-medium ${getStateColor(
                          build.state
                        )}`}
                      >
                        {build.state}
                      </span>
                    </div>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      Target: {build.target}
                    </p>
                    <p className="mt-1 text-xs text-gray-500 dark:text-gray-500">
                      Started: {new Date(build.startTime).toLocaleString()}
                    </p>
                  </div>
                  <svg
                    className="h-6 w-6 text-gray-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 5l7 7-7 7"
                    />
                  </svg>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
