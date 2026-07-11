"use client";

import { ArrowRight, Brain, Shield, Search, Moon, Bell, Check, Menu, X, Slack, Github, BookOpen, Star } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { ThemeToggle } from "@/components/ThemeToggle";

const features = [
  {
    icon: Brain,
    title: "Focus Mode",
    desc: "Detects fast-moving channels and generates AI decision trees so you never miss what matters.",
    color: "from-violet-500 to-purple-600",
    bgColor: "bg-violet-50 dark:bg-violet-950/50",
  },
  {
    icon: Shield,
    title: "Social Translator",
    desc: "Decodes passive-aggressive and ambiguous messages into plain, direct language via DM.",
    color: "from-emerald-500 to-teal-600",
    bgColor: "bg-emerald-50 dark:bg-emerald-950/50",
  },
  {
    icon: Search,
    title: "Catch-Up",
    desc: "Ask What did I miss about {topic}? and get a structured AI summary from Slack search.",
    color: "from-blue-500 to-indigo-600",
    bgColor: "bg-blue-50 dark:bg-blue-950/50",
  },
  {
    icon: Moon,
    title: "Deep Work Protector",
    desc: "Blocks focus time on your calendar, sets Slack status, and pauses notifications. Powered by MCP.",
    color: "from-amber-500 to-orange-600",
    bgColor: "bg-amber-50 dark:bg-amber-950/50",
  },
  {
    icon: Bell,
    title: "Quiet Hours Digest",
    desc: "Batches non-urgent mentions into a structured 4 PM digest. No more notification overload.",
    color: "from-rose-500 to-pink-600",
    bgColor: "bg-rose-50 dark:bg-rose-950/50",
  },
  {
    icon: Star,
    title: "Personalized for You",
    desc: "Configure threshold, neurotype, digest time, and more. Signal adapts to your brain.",
    color: "from-cyan-500 to-sky-600",
    bgColor: "bg-cyan-50 dark:bg-cyan-950/50",
  },
];

const personas = [
  {
    name: "Alex",
    neurotype: "ADHD",
    role: "Software Engineer",
    quote: "I'd return from lunch to 200 messages. Now Focus Mode finds the decisions for me.",
    color: "bg-violet-100 dark:bg-violet-950 text-violet-800 dark:text-violet-200",
  },
  {
    name: "Jordan",
    neurotype: "Autistic",
    role: "Product Manager",
    quote: "Per my last email used to ruin my afternoon. Signal tells me what they actually meant.",
    color: "bg-emerald-100 dark:bg-emerald-950 text-emerald-800 dark:text-emerald-200",
  },
  {
    name: "Taylor",
    neurotype: "ADHD",
    role: "Designer",
    quote: "Deep Work mode saved my focus. No more 23-min cycles broken by @mentions.",
    color: "bg-amber-100 dark:bg-amber-950 text-amber-800 dark:text-amber-200",
  },
];

const stats = [
  { value: "15–20%", label: "of workforce is neurodivergent" },
  { value: "Zero", label: "accessibility tools in Slack Marketplace" },
  { value: "3 APIs", label: "Slack AI + MCP + Real-Time Search" },
  { value: "6 Days", label: "from idea to working prototype" },
];

export default function Home() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <div className="flex min-h-screen flex-col">
      {/* Navigation */}
      <header className="sticky top-0 z-50 border-b border-zinc-200/50 dark:border-zinc-800/50 bg-white/80 dark:bg-zinc-950/80 backdrop-blur-xl">
        <nav className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-signal-blue to-signal-teal shadow-sm">
              <span className="text-sm font-bold text-white">S</span>
            </div>
            <span className="text-lg font-bold tracking-tight">Signal</span>
          </div>

          <div className="hidden md:flex items-center gap-8">
            <a href="#features" className="text-sm font-medium text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100 transition-colors">
              Features
            </a>
            <a href="#how-it-works" className="text-sm font-medium text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100 transition-colors">
              How It Works
            </a>
            <a href="#stories" className="text-sm font-medium text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100 transition-colors">
              Stories
            </a>
            <a href="/docs" className="text-sm font-medium text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100 transition-colors">
              Docs
            </a>
          </div>

          <div className="flex items-center gap-3">
            <ThemeToggle />
            <a
              href="https://slack.com/oauth/v2/authorize"
              className="hidden sm:inline-flex items-center gap-2 rounded-lg bg-signal-blue px-5 py-2.5 text-sm font-semibold text-white shadow-sm transition-all hover:bg-signal-blue-dark hover:shadow-md"
            >
              <Slack className="h-4 w-4" />
              Add to Slack
            </a>
            <button
              className="md:hidden p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800"
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            >
              {mobileMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
            </button>
          </div>
        </nav>

        {/* Mobile menu */}
        {mobileMenuOpen && (
          <div className="md:hidden border-t border-zinc-200 dark:border-zinc-800 px-6 py-4 bg-white dark:bg-zinc-950">
            <div className="flex flex-col gap-3">
              <a href="#features" className="text-sm font-medium py-2" onClick={() => setMobileMenuOpen(false)}>Features</a>
              <a href="#how-it-works" className="text-sm font-medium py-2" onClick={() => setMobileMenuOpen(false)}>How It Works</a>
              <a href="#stories" className="text-sm font-medium py-2" onClick={() => setMobileMenuOpen(false)}>Stories</a>
              <a href="/docs" className="text-sm font-medium py-2" onClick={() => setMobileMenuOpen(false)}>Docs</a>
              <a
                href="https://slack.com/oauth/v2/authorize"
                className="inline-flex items-center justify-center gap-2 rounded-lg bg-signal-blue px-5 py-2.5 text-sm font-semibold text-white mt-2"
              >
                <Slack className="h-4 w-4" />
                Add to Slack
              </a>
            </div>
          </div>
        )}
      </header>

      <main className="flex-1">
        {/* Hero Section */}
        <section className="relative overflow-hidden px-6 py-24 sm:py-32 lg:py-40">
          <div className="absolute inset-0 -z-10 bg-[linear-gradient(to_right,#80808012_1px,transparent_1px),linear-gradient(to_bottom,#80808012_1px,transparent_1px)] bg-[size:32px_32px]" />
          <div className="absolute -top-40 -right-40 h-80 w-80 rounded-full bg-signal-blue/10 blur-3xl" />
          <div className="absolute -bottom-40 -left-40 h-80 w-80 rounded-full bg-signal-teal/10 blur-3xl" />

          <div className="mx-auto max-w-4xl text-center">
            <div className="inline-flex items-center gap-2 rounded-full border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 px-4 py-1.5 text-sm mb-8 shadow-sm">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-signal-teal opacity-75" />
                <span className="relative inline-flex rounded-full h-2 w-2 bg-signal-teal" />
              </span>
              Slack Agent Builder Challenge — Submission
            </div>

            <h1 className="text-4xl font-bold tracking-tight sm:text-6xl lg:text-7xl">
              Slack is overwhelming.
              <br />
              <span className="gradient-text">Signal makes it calm.</span>
            </h1>

            <p className="mt-6 text-lg leading-8 text-zinc-600 dark:text-zinc-400 max-w-2xl mx-auto">
              The first Slack agent purpose-built for neurodivergent professionals.
              Focus Mode, Social Translator, Catch-Up, and Deep Work Protector —
              all powered by Slack AI, MCP, and Real-Time Search.
            </p>

            <div className="mt-10 flex flex-col sm:flex-row items-center justify-center gap-4">
              <a
                href="https://slack.com/oauth/v2/authorize"
                className="btn-primary text-base px-8 py-3.5 w-full sm:w-auto"
              >
                <Slack className="h-5 w-5" />
                Add to Slack
                <ArrowRight className="h-4 w-4" />
              </a>
              <a
                href="#features"
                className="btn-secondary text-base px-8 py-3.5 w-full sm:w-auto"
              >
                See How It Works
              </a>
            </div>

            <div className="mt-16 grid grid-cols-2 gap-4 sm:grid-cols-4">
              {stats.map((stat) => (
                <div key={stat.label} className="glass rounded-xl p-4 text-center">
                  <div className="text-2xl font-bold gradient-text">{stat.value}</div>
                  <div className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">{stat.label}</div>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Features Grid */}
        <section id="features" className="px-6 py-24 bg-zinc-50/50 dark:bg-zinc-900/50">
          <div className="mx-auto max-w-7xl">
            <div className="text-center mb-16">
              <span className="tag mb-4">Features</span>
              <h2 className="text-3xl font-bold tracking-tight sm:text-4xl">
                Everything you need to
                <br />
                <span className="gradient-text">tame the chaos</span>
              </h2>
              <p className="mt-4 text-zinc-600 dark:text-zinc-400 max-w-xl mx-auto">
                Five integrated features that work together to make Slack accessible for every brain type.
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {features.map((feature) => (
                <div key={feature.title} className="feature-card group">
                  <div className={`inline-flex h-12 w-12 items-center justify-center rounded-xl ${feature.bgColor} mb-4`}>
                    <feature.icon className={`h-6 w-6 bg-gradient-to-br ${feature.color} bg-clip-text text-transparent`} />
                  </div>
                  <h3 className="text-lg font-semibold mb-2">{feature.title}</h3>
                  <p className="text-sm text-zinc-600 dark:text-zinc-400 leading-relaxed">
                    {feature.desc}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* How It Works */}
        <section id="how-it-works" className="px-6 py-24">
          <div className="mx-auto max-w-5xl">
            <div className="text-center mb-16">
              <span className="tag mb-4">How It Works</span>
              <h2 className="text-3xl font-bold tracking-tight sm:text-4xl">
                Install. Configure. <span className="gradient-text">Relax.</span>
              </h2>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
              {[
                { step: "01", title: "Install", desc: "Add Signal to your Slack workspace with one click. No complex setup." },
                { step: "02", title: "Configure", desc: "Set your neurotype, preferences, and thresholds. Signal learns your style." },
                { step: "03", title: "Relax", desc: "Signal works automatically. Focus Mode, Translator, Catch-Up — all passive." },
              ].map((item) => (
                <div key={item.step} className="text-center">
                  <div className="inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-signal-blue to-signal-teal text-white text-xl font-bold mb-6 shadow-lg">
                    {item.step}
                  </div>
                  <h3 className="text-lg font-semibold mb-2">{item.title}</h3>
                  <p className="text-sm text-zinc-600 dark:text-zinc-400 max-w-xs mx-auto">{item.desc}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* User Stories */}
        <section id="stories" className="px-6 py-24 bg-zinc-50/50 dark:bg-zinc-900/50">
          <div className="mx-auto max-w-5xl">
            <div className="text-center mb-16">
              <span className="tag mb-4">Real People</span>
              <h2 className="text-3xl font-bold tracking-tight sm:text-4xl">
                Built for <span className="gradient-text">real brains</span>
              </h2>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {personas.map((persona) => (
                <div key={persona.name} className="glass rounded-2xl p-6 card-hover">
                  <div className="flex items-center gap-3 mb-4">
                    <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-signal-blue to-signal-teal text-white font-bold text-sm">
                      {persona.name[0]}
                    </div>
                    <div>
                      <div className="font-semibold text-sm">{persona.name}</div>
                      <div className="text-xs text-zinc-500">{persona.role}</div>
                    </div>
                  </div>
                  <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium mb-3 ${persona.color}`}>
                    {persona.neurotype}
                  </span>
                  <p className="text-sm text-zinc-600 dark:text-zinc-400 italic leading-relaxed">
                    &ldquo;{persona.quote}&rdquo;
                  </p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* CTA Section */}
        <section className="px-6 py-24">
          <div className="mx-auto max-w-3xl text-center">
            <div className="glass rounded-3xl p-12 shadow-2xl">
              <Brain className="h-12 w-12 mx-auto text-signal-blue mb-6" />
              <h2 className="text-3xl font-bold tracking-tight sm:text-4xl mb-4">
                Ready for <span className="gradient-text">calmer</span> Slack?
              </h2>
              <p className="text-lg text-zinc-600 dark:text-zinc-400 mb-8 max-w-lg mx-auto">
                Join the first neurodivergent accessibility tool for Slack. Free during hackathon.
              </p>
              <a
                href="https://slack.com/oauth/v2/authorize"
                className="btn-primary text-base px-8 py-3.5"
              >
                <Slack className="h-5 w-5" />
                Add Signal to Slack
              </a>
            </div>
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="border-t border-zinc-200 dark:border-zinc-800 px-6 py-12">
        <div className="mx-auto max-w-7xl">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
            <div className="col-span-2 md:col-span-1">
              <div className="flex items-center gap-2 mb-4">
                <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-gradient-to-br from-signal-blue to-signal-teal">
                  <span className="text-xs font-bold text-white">S</span>
                </div>
                <span className="font-bold">Signal</span>
              </div>
              <p className="text-sm text-zinc-500 dark:text-zinc-400">
                Making Slack accessible for everyone.
              </p>
            </div>
            <div>
              <h4 className="font-semibold text-sm mb-3">Product</h4>
              <div className="flex flex-col gap-2 text-sm text-zinc-500 dark:text-zinc-400">
                <a href="#features" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">Features</a>
                <a href="#how-it-works" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">How It Works</a>
                <a href="/docs" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">Documentation</a>
              </div>
            </div>
            <div>
              <h4 className="font-semibold text-sm mb-3">Resources</h4>
              <div className="flex flex-col gap-2 text-sm text-zinc-500 dark:text-zinc-400">
                <a href="https://github.com/LSUDOKOS/signal" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors flex items-center gap-1">
                  <Github className="h-3.5 w-3.5" /> GitHub
                </a>
                <a href="/privacy" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">Privacy Policy</a>
                <a href="/support" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">Support</a>
              </div>
            </div>
            <div>
              <h4 className="font-semibold text-sm mb-3">Hackathon</h4>
              <div className="flex flex-col gap-2 text-sm text-zinc-500 dark:text-zinc-400">
                <a href="https://devpost.com" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">Devpost</a>
                <a href="https://api.slack.com" className="hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">Slack API</a>
                <span className="text-xs">Slack Agent for Good</span>
              </div>
            </div>
          </div>
          <div className="mt-8 pt-8 border-t border-zinc-200 dark:border-zinc-800 text-center text-sm text-zinc-500">
            Built with ♥ for the Slack Agent Builder Challenge • July 2026
          </div>
        </div>
      </footer>
    </div>
  );
}
