import Link from "next/link";

export default function Home() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800">
      <div className="container mx-auto px-4 py-16">
        <header className="mb-12 text-center">
          <h1 className="mb-4 text-5xl font-bold text-gray-900 dark:text-white">
            Smidr
          </h1>
          <p className="text-xl text-gray-600 dark:text-gray-300">
            Yocto Build Management System
          </p>
        </header>

        <main className="mx-auto max-w-4xl">
          <div className="mb-8 rounded-lg bg-white p-8 shadow-lg dark:bg-gray-800">
            <h2 className="mb-4 text-2xl font-semibold text-gray-900 dark:text-white">
              Welcome to Smidr Web Interface
            </h2>
            <p className="mb-6 text-gray-700 dark:text-gray-300">
              Manage your Yocto builds through an intuitive web interface.
              Monitor build status, view logs, and download artifacts all in one
              place.
            </p>

            <div className="grid gap-4 md:grid-cols-2">
              <Link
                href="/builds"
                className="flex flex-col items-center rounded-lg border border-gray-200 bg-gray-50 p-6 transition-all hover:border-blue-500 hover:bg-blue-50 dark:border-gray-700 dark:bg-gray-900 dark:hover:border-blue-400 dark:hover:bg-gray-700"
              >
                <svg
                  className="mb-3 h-12 w-12 text-blue-600 dark:text-blue-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                  />
                </svg>
                <h3 className="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
                  View Builds
                </h3>
                <p className="text-center text-sm text-gray-600 dark:text-gray-400">
                  Browse all builds and their status
                </p>
              </Link>

              <Link
                href="/builds/new"
                className="flex flex-col items-center rounded-lg border border-gray-200 bg-gray-50 p-6 transition-all hover:border-green-500 hover:bg-green-50 dark:border-gray-700 dark:bg-gray-900 dark:hover:border-green-400 dark:hover:bg-gray-700"
              >
                <svg
                  className="mb-3 h-12 w-12 text-green-600 dark:text-green-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 4v16m8-8H4"
                  />
                </svg>
                <h3 className="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
                  Start New Build
                </h3>
                <p className="text-center text-sm text-gray-600 dark:text-gray-400">
                  Configure and launch a new build
                </p>
              </Link>
            </div>
          </div>

          <div className="rounded-lg bg-white p-8 shadow-lg dark:bg-gray-800">
            <h2 className="mb-4 text-xl font-semibold text-gray-900 dark:text-white">
              Features
            </h2>
            <ul className="space-y-3 text-gray-700 dark:text-gray-300">
              <li className="flex items-start">
                <svg
                  className="mr-2 mt-1 h-5 w-5 flex-shrink-0 text-green-500"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                    clipRule="evenodd"
                  />
                </svg>
                <span>Monitor build status in real-time</span>
              </li>
              <li className="flex items-start">
                <svg
                  className="mr-2 mt-1 h-5 w-5 flex-shrink-0 text-green-500"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                    clipRule="evenodd"
                  />
                </svg>
                <span>Stream build logs live</span>
              </li>
              <li className="flex items-start">
                <svg
                  className="mr-2 mt-1 h-5 w-5 flex-shrink-0 text-green-500"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                    clipRule="evenodd"
                  />
                </svg>
                <span>Download build artifacts</span>
              </li>
              <li className="flex items-start">
                <svg
                  className="mr-2 mt-1 h-5 w-5 flex-shrink-0 text-green-500"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                    clipRule="evenodd"
                  />
                </svg>
                <span>Manage multiple concurrent builds</span>
              </li>
            </ul>
          </div>
        </main>
      </div>
    </div>
  );
}
