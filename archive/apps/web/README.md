# Smidr Web Application

A modern Next.js web interface for the Smidr Yocto build management system.

## Features

- 🔍 **View Builds**: Browse all builds with real-time status updates
- ▶️ **Start Builds**: Configure and launch new Yocto builds
- 📊 **Build Details**: Monitor build progress and view detailed information
- 📝 **Live Logs**: Stream build logs in real-time (coming soon)
- 📦 **Artifacts**: Download build artifacts (coming soon)

## Tech Stack

- **Next.js 16** with App Router
- **TypeScript** for type safety
- **Tailwind CSS** for styling
- **React 19** for UI components

## Getting Started

First, make sure the Smidr API server is running (default: `http://localhost:5285`).

Then run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

## Configuration

Configure the API endpoint in `.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:5285
```

## Project Structure

```
app/
├── page.tsx              # Home page
├── builds/
│   ├── page.tsx         # Builds list
│   ├── new/
│   │   └── page.tsx     # New build form
│   └── [buildId]/
│       └── page.tsx     # Build details
├── layout.tsx           # Root layout
└── globals.css          # Global styles
```

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.
