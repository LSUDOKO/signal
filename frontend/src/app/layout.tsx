import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { Providers } from "./providers";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
});

export const metadata: Metadata = {
  title: "Signal — Calm Slack for Neurodivergent Professionals",
  description:
    "Signal transforms overwhelming Slack experiences into calm, structured, and comprehensible interactions for ADHD, autistic, and anxious professionals.",
  openGraph: {
    title: "Signal — Calm Slack for Neurodivergent Professionals",
    description:
      "Focus Mode, Social Translator, Catch-Up, and Deep Work Protector — all inside Slack.",
    type: "website",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${inter.variable} h-full antialiased`}
      suppressHydrationWarning
    >
      <body className="min-h-full flex flex-col bg-white text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
