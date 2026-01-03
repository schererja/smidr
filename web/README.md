# Smidr Web Frontend

Modern Next.js/React frontend for the Smidr CI/CD platform.

## Prerequisites

- Node.js 18+
- The Smidr backend server running (default: <http://localhost:8080>)

## Getting Started

First, install dependencies:

```bash
cd web
npm install
```

Copy the environment configuration:

```bash
cp .env.example .env.local
```

Then, run the development server:

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

## Configuration

The frontend connects to the backend API via the `NEXT_PUBLIC_API_URL` environment variable:

```bash
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Pages

- **Home** (`/`) - Landing page with feature overview
- **Dashboard** (`/dashboard`) - Overview of connected agents and system status
- **Agents** (`/agents`) - List of all registered agents
- **Agent Details** (`/agents/[id]`) - Detailed information about a specific agent

## Tech Stack

- **Next.js 14** - React framework with App Router
- **TypeScript** - Type-safe JavaScript
- **Tailwind CSS** - Utility-first CSS framework
- **React 18** - UI library

## Project Structure

```
web/
├── app/                  # Next.js App Router
│   ├── layout.tsx       # Root layout
│   ├── page.tsx         # Home page
│   ├── dashboard/       # Dashboard pages
│   ├── agents/          # Agent pages
│   └── globals.css      # Global styles
├── lib/
│   ├── api.ts           # API client
│   └── hooks.ts         # React hooks for data fetching
├── components/          # Reusable React components
├── public/             # Static assets
└── package.json        # Dependencies
```

## API Integration

The frontend connects to the Smidr backend using these endpoints:

- `GET /health` - Health check
- `GET /api/v1/agents` - List all agents
- `GET /api/v1/agents/{agentID}` - Get agent details
- `GET /api/v1/jobs` - List all jobs
- `POST /api/v1/jobs` - Create a new job
- `GET /api/v1/jobs/{jobID}` - Get job details

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm start` - Start production server
- `npm run lint` - Run ESLint

## Learn More

- [Next.js Documentation](https://nextjs.org/docs)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [React Documentation](https://react.dev)
