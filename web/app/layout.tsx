import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Smidr - CI/CD Platform",
  description: "Modern CI/CD platform for building and deploying applications",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="antialiased">{children}</body>
    </html>
  );
}
