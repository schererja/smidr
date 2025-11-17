"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";

interface BuildStatus {
  buildId: string;
  state: string;
  startTime: string;
  endTime?: string;
  errorMessage?: string;
  exitCode: number;
}

export default function BuildDetailPage() {
  const params = useParams();
  const buildId = params.buildId as string;

  const [build, setBuild] = useState<BuildStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchBuild = async () => {
      try {
        const apiUrl =
          process.env.NEXT_PUBLIC_API_URL || "http://localhost:5285";
        const response = await fetch(`${apiUrl}/api/builds/${buildId}`);

        if (!response.ok) {
          throw new Error(`Failed to fetch build: ${response.statusText}`);
        }

        const data = await response.json();
        setBuild(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to fetch build");
      } finally {
        setLoading(false);
      }
    };

    fetchBuild();

    // Refresh every 3 seconds if build is in progress
    const interval = setInterval(() => {
      if (
        build &&
        !["Completed", "Failed", "Cancelled"].includes(build.state)
      ) {
        fetchBuild();
      }
    }, 3000);

    return () => clearInterval(interval);
  }, [buildId, build]);

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
        <div className="mb-8">
          <Link
            href="/builds"
            className="mb-2 inline-block text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400"
          >
            ← Back to Builds
          </Link>
          <h1 className="text-4xl font-bold text-gray-900 dark:text-white">
            Build Details
          </h1>
        </div>

        {loading && (
          <div className="rounded-lg bg-white p-8 text-center shadow-lg dark:bg-gray-800">
            <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-blue-600 border-r-transparent"></div>
            <p className="mt-4 text-gray-600 dark:text-gray-400">
              Loading build...
            </p>
          </div>
        )}

        {error && (
          <div className="rounded-lg bg-red-50 p-4 text-red-800 dark:bg-red-900/20 dark:text-red-200">
            <p className="font-semibold">Error loading build</p>
            <p className="text-sm">{error}</p>
          </div>
        )}

        {!loading && !error && build && (
          <div className="space-y-6">
            <div className="rounded-lg bg-white p-6 shadow-lg dark:bg-gray-800">
              <div className="mb-4 flex items-center justify-between">
                <h2 className="text-2xl font-semibold text-gray-900 dark:text-white">
                  {build.buildId}
                </h2>
                <span
                  className={`rounded-full px-4 py-2 text-sm font-medium ${getStateColor(
                    build.state
                  )}`}
                >
                  {build.state}
                </span>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">
                    Start Time
                  </p>
                  <p className="text-gray-900 dark:text-white">
                    {new Date(build.startTime).toLocaleString()}
                  </p>
                </div>

                {build.endTime && (
                  <div>
                    <p className="text-sm font-medium text-gray-600 dark:text-gray-400">
                      End Time
                    </p>
                    <p className="text-gray-900 dark:text-white">
                      {new Date(build.endTime).toLocaleString()}
                    </p>
                  </div>
                )}

                <div>
                  <p className="text-sm font-medium text-gray-600 dark:text-gray-400">
                    Exit Code
                  </p>
                  <p className="text-gray-900 dark:text-white">
                    {build.exitCode}
                  </p>
                </div>
              </div>

              {build.errorMessage && (
                <div className="mt-4 rounded-lg bg-red-50 p-4 dark:bg-red-900/20">
                  <p className="text-sm font-medium text-red-800 dark:text-red-200">
                    Error Message
                  </p>
                  <p className="mt-1 text-sm text-red-700 dark:text-red-300">
                    {build.errorMessage}
                  </p>
                </div>
              )}
            </div>

            <div className="rounded-lg bg-white p-6 shadow-lg dark:bg-gray-800">
              <h3 className="mb-4 text-xl font-semibold text-gray-900 dark:text-white">
                Actions
              </h3>
              <div className="flex flex-wrap gap-3">
                <Link
                  href={`/builds/${buildId}/logs`}
                  className="rounded-lg bg-blue-600 px-4 py-2 text-white transition-colors hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600"
                >
                  View Logs
                </Link>
                <Link
                  href={`/builds/${buildId}/artifacts`}
                  className="rounded-lg bg-green-600 px-4 py-2 text-white transition-colors hover:bg-green-700 dark:bg-green-500 dark:hover:bg-green-600"
                >
                  View Artifacts
                </Link>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
