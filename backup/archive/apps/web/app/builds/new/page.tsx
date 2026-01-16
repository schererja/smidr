"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";

export default function NewBuildPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [formData, setFormData] = useState({
    configPath: "",
    target: "core-image-minimal",
    customer: "",
    forceClean: false,
    forceImageRebuild: false,
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:5285";
      const response = await fetch(`${apiUrl}/api/builds`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(formData),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message || `Failed to start build: ${response.statusText}`
        );
      }

      const data = await response.json();
      router.push(`/builds/${data.buildId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to start build");
      setLoading(false);
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
            Start New Build
          </h1>
        </div>

        <div className="mx-auto max-w-2xl rounded-lg bg-white p-8 shadow-lg dark:bg-gray-800">
          {error && (
            <div className="mb-6 rounded-lg bg-red-50 p-4 text-red-800 dark:bg-red-900/20 dark:text-red-200">
              <p className="font-semibold">Error</p>
              <p className="text-sm">{error}</p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label
                htmlFor="configPath"
                className="mb-2 block text-sm font-medium text-gray-900 dark:text-white"
              >
                Config Path <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                id="configPath"
                required
                value={formData.configPath}
                onChange={(e) =>
                  setFormData({ ...formData, configPath: e.target.value })
                }
                className="w-full rounded-lg border border-gray-300 bg-white px-4 py-2 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
                placeholder="/path/to/smidr.yaml"
              />
              <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
                Path to your smidr configuration file
              </p>
            </div>

            <div>
              <label
                htmlFor="target"
                className="mb-2 block text-sm font-medium text-gray-900 dark:text-white"
              >
                Build Target <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                id="target"
                required
                value={formData.target}
                onChange={(e) =>
                  setFormData({ ...formData, target: e.target.value })
                }
                className="w-full rounded-lg border border-gray-300 bg-white px-4 py-2 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
                placeholder="core-image-minimal"
              />
              <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
                Yocto image target to build
              </p>
            </div>

            <div>
              <label
                htmlFor="customer"
                className="mb-2 block text-sm font-medium text-gray-900 dark:text-white"
              >
                Customer (Optional)
              </label>
              <input
                type="text"
                id="customer"
                value={formData.customer}
                onChange={(e) =>
                  setFormData({ ...formData, customer: e.target.value })
                }
                className="w-full rounded-lg border border-gray-300 bg-white px-4 py-2 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
                placeholder="customer-name"
              />
            </div>

            <div className="space-y-3">
              <div className="flex items-center">
                <input
                  type="checkbox"
                  id="forceClean"
                  checked={formData.forceClean}
                  onChange={(e) =>
                    setFormData({ ...formData, forceClean: e.target.checked })
                  }
                  className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-2 focus:ring-blue-500"
                />
                <label
                  htmlFor="forceClean"
                  className="ml-2 text-sm text-gray-900 dark:text-white"
                >
                  Force Clean Build
                </label>
              </div>

              <div className="flex items-center">
                <input
                  type="checkbox"
                  id="forceImageRebuild"
                  checked={formData.forceImageRebuild}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      forceImageRebuild: e.target.checked,
                    })
                  }
                  className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-2 focus:ring-blue-500"
                />
                <label
                  htmlFor="forceImageRebuild"
                  className="ml-2 text-sm text-gray-900 dark:text-white"
                >
                  Force Image Rebuild
                </label>
              </div>
            </div>

            <div className="flex gap-4">
              <button
                type="submit"
                disabled={loading}
                className="flex-1 rounded-lg bg-blue-600 px-6 py-3 font-medium text-white transition-colors hover:bg-blue-700 disabled:cursor-not-allowed disabled:bg-gray-400 dark:bg-blue-500 dark:hover:bg-blue-600"
              >
                {loading ? "Starting Build..." : "Start Build"}
              </button>
              <Link
                href="/builds"
                className="rounded-lg border border-gray-300 px-6 py-3 font-medium text-gray-700 transition-colors hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
              >
                Cancel
              </Link>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
