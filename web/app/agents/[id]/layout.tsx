import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Agent Details - Smidr",
  description: "View detailed information about a Smidr agent",
};

export default function Layout({ children }: { children: React.ReactNode }) {
  return children;
}
