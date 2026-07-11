import Link from "next/link";

export default function Privacy() {
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
        <h1 className="text-3xl font-bold mb-8">Privacy Policy</h1>
        <p className="text-sm text-zinc-500 mb-8">Last updated: July 11, 2026</p>

        <div className="prose prose-zinc dark:prose-invert max-w-none">
          <h2>Data We Collect</h2>
          <ul>
            <li><strong>Slack User ID & Team ID:</strong> Required to associate preferences and send DMs.</li>
            <li><strong>User Preferences:</strong> Neurotype, feature toggles, threshold settings.</li>
            <li><strong>Message Content:</strong> Transiently processed for Focus Mode summaries and Social Translator. <strong>Never permanently stored.</strong></li>
            <li><strong>Channel Metadata:</strong> Channel IDs and names for velocity tracking and subscriptions.</li>
          </ul>

          <h2>Data We Do NOT Collect</h2>
          <ul>
            <li>We do not store message content permanently.</li>
            <li>We do not sell user data.</li>
            <li>We do not share personal information with third parties.</li>
            <li>We do not track or fingerprint users across workspaces.</li>
          </ul>

          <h2>Data Retention</h2>
          <p>Message content is processed in-memory and discarded immediately after AI analysis. User preferences are stored until the app is uninstalled or you delete your account.</p>

          <h2>Third-Party Services</h2>
          <p>Signal uses OpenAI&apos;s API (GPT-4o-mini) for AI analysis. Messages sent to OpenAI are not used for training. See <a href="https://openai.com/policies/privacy-policy">OpenAI&apos;s Privacy Policy</a>.</p>

          <h2>Contact</h2>
          <p>For privacy concerns, open an issue on <a href="https://github.com/LSUDOKOS/signal">GitHub</a> or email privacy@signal-slack.app.</p>
        </div>
      </main>
    </div>
  );
}
