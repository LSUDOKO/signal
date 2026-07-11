import Link from "next/link";
import { GitFork, BookOpen, MessageCircle } from "lucide-react";

export default function Support() {
  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950">
      <header className="border-b border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900">
        <div className="mx-auto max-w-3xl px-6 py-4 flex items-center gap-3">
          <Link href="/" className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-signal-blue to-signal-teal shadow-sm">
              <span className="text-sm font-bold text-white">S</span>
            </div>
            <span className="font-semibold">Signal</span>
          </Link>
        </div>
      </header>

      <main className="mx-auto max-w-3xl px-6 py-16">
        <h1 className="text-3xl font-bold mb-4">Support</h1>
        <p className="text-zinc-500 dark:text-zinc-400 mb-12">
          Need help with Signal? Here are the best ways to get it.
        </p>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
          <a
            href="https://github.com/LSUDOKOS/signal"
            className="glass rounded-2xl p-6 card-hover text-center"
          >
            <GitFork className="h-8 w-8 mx-auto mb-4 text-signal-blue" />
            <h3 className="font-semibold mb-2">GitHub Issues</h3>
            <p className="text-sm text-zinc-500">Report bugs or request features</p>
          </a>

          <a
            href="/docs"
            className="glass rounded-2xl p-6 card-hover text-center"
          >
            <BookOpen className="h-8 w-8 mx-auto mb-4 text-signal-teal" />
            <h3 className="font-semibold mb-2">Documentation</h3>
            <p className="text-sm text-zinc-500">Read the full documentation</p>
          </a>

          <a
            href="https://github.com/LSUDOKOS/signal/discussions"
            className="glass rounded-2xl p-6 card-hover text-center"
          >
            <MessageCircle className="h-8 w-8 mx-auto mb-4 text-signal-amber" />
            <h3 className="font-semibold mb-2">Discussions</h3>
            <p className="text-sm text-zinc-500">Join the community conversation</p>
          </a>
        </div>
      </main>
    </div>
  );
}
