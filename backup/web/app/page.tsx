import Link from "next/link";

export default function Home() {
  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b">
        <nav className="container mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <div className="text-2xl font-bold text-blue-600">Smidr</div>
          </div>
          <div className="flex items-center space-x-6">
            <Link href="/docs" className="text-gray-600 hover:text-gray-900">
              Docs
            </Link>
            <Link
              href="/dashboard"
              className="text-gray-600 hover:text-gray-900"
            >
              Dashboard
            </Link>
            <Link href="/agents" className="text-gray-600 hover:text-gray-900">
              Agents
            </Link>
            <Link
              href="/dashboard"
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
            >
              Launch
            </Link>
          </div>
        </nav>
      </header>

      <main className="flex-1">
        <section className="container mx-auto px-4 py-20 text-center">
          <h1 className="text-5xl font-bold mb-6">Modern CI/CD Platform</h1>
          <p className="text-xl text-gray-600 mb-8 max-w-2xl mx-auto">
            Build, test, and deploy your applications with ease. Smidr provides
            a powerful and flexible CI/CD solution for modern development teams.
          </p>
          <div className="flex gap-4 justify-center">
            <Link
              href="/getting-started"
              className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition text-lg font-medium"
            >
              Get Started
            </Link>
            <Link
              href="/docs"
              className="px-6 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition text-lg font-medium"
            >
              Learn More
            </Link>
          </div>
        </section>

        <section className="bg-gray-50 py-16">
          <div className="container mx-auto px-4">
            <h2 className="text-3xl font-bold text-center mb-12">Features</h2>
            <div className="grid md:grid-cols-3 gap-8">
              <div className="bg-white p-6 rounded-lg shadow-sm">
                <div className="text-blue-600 text-3xl mb-4">⚡</div>
                <h3 className="text-xl font-semibold mb-2">Fast Builds</h3>
                <p className="text-gray-600">
                  Optimized build pipelines with caching and parallel execution
                  for maximum speed.
                </p>
              </div>
              <div className="bg-white p-6 rounded-lg shadow-sm">
                <div className="text-blue-600 text-3xl mb-4">🔧</div>
                <h3 className="text-xl font-semibold mb-2">
                  Flexible Configuration
                </h3>
                <p className="text-gray-600">
                  YAML-based configuration with support for custom workflows and
                  integrations.
                </p>
              </div>
              <div className="bg-white p-6 rounded-lg shadow-sm">
                <div className="text-blue-600 text-3xl mb-4">📊</div>
                <h3 className="text-xl font-semibold mb-2">
                  Real-time Monitoring
                </h3>
                <p className="text-gray-600">
                  Track your builds and deployments with detailed logs and
                  status updates.
                </p>
              </div>
            </div>
          </div>
        </section>
      </main>

      <footer className="border-t py-8">
        <div className="container mx-auto px-4 text-center text-gray-600">
          <p>&copy; 2026 Smidr. All rights reserved.</p>
        </div>
      </footer>
    </div>
  );
}
