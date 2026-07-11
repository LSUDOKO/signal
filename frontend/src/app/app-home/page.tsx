"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { Brain, Save, ArrowLeft, Zap, Shuffle, Bell, Shield, Moon } from "lucide-react";
import Link from "next/link";

const prefsSchema = z.object({
  neurotype: z.enum(["adhd", "autism", "anxiety", "unspecified", "ally"]),
  focus_mode_enabled: z.boolean(),
  focus_threshold: z.number().min(10).max(100),
  translator_enabled: z.boolean(),
  digest_enabled: z.boolean(),
  digest_hour: z.number().min(0).max(23),
  deep_work_auto_detect: z.boolean(),
  quiet_hours_start: z.string(),
  quiet_hours_end: z.string(),
});

type PrefsForm = z.infer<typeof prefsSchema>;

const neurotypes = [
  { value: "adhd", label: "ADHD", desc: "Prioritize focus assistance and digest batching" },
  { value: "autism", label: "Autistic", desc: "Prioritize social translation and clear language" },
  { value: "anxiety", label: "Anxiety", desc: "Prioritize quiet hours and gentle notifications" },
  { value: "unspecified", label: "Unsure", desc: "Balanced defaults, adapt to your usage" },
  { value: "ally", label: "Ally / Manager", desc: "Enable features for your team members" },
];

export default function AppHome() {
  const [saving, setSaving] = useState(false);

  const form = useForm<PrefsForm>({
    resolver: zodResolver(prefsSchema),
    defaultValues: {
      neurotype: "unspecified",
      focus_mode_enabled: true,
      focus_threshold: 50,
      translator_enabled: true,
      digest_enabled: false,
      digest_hour: 16,
      deep_work_auto_detect: false,
      quiet_hours_start: "22:00",
      quiet_hours_end: "08:00",
    },
  });

  const watchNeurotype = form.watch("neurotype");
  const watchDigest = form.watch("digest_enabled");

  const onSubmit = async (data: PrefsForm) => {
    setSaving(true);
    // Simulate API call
    await new Promise((r) => setTimeout(r, 800));
    toast.success("Preferences saved! Signal will adapt to your settings.");
    setSaving(false);
  };

  return (
    <div className="min-h-screen bg-zinc-50 dark:bg-zinc-950">
      <header className="border-b border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900">
        <div className="mx-auto max-w-3xl px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Link href="/" className="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors">
              <ArrowLeft className="h-5 w-5" />
            </Link>
            <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-gradient-to-br from-signal-blue to-signal-teal shadow-sm">
              <span className="text-sm font-bold text-white">S</span>
            </div>
            <span className="font-semibold">Signal Preferences</span>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-3xl px-6 py-12">
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-10">
          {/* Neurotype */}
          <section>
            <div className="flex items-center gap-2 mb-6">
              <Brain className="h-5 w-5 text-signal-blue" />
              <h2 className="text-lg font-semibold">Your Neurotype</h2>
            </div>
            <p className="text-sm text-zinc-500 dark:text-zinc-400 mb-4">
              This helps Signal tailor its features to your needs. Change anytime.
            </p>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {neurotypes.map((nt) => (
                <label
                  key={nt.value}
                  className={`relative flex cursor-pointer rounded-xl border-2 p-4 transition-all ${
                    watchNeurotype === nt.value
                      ? "border-signal-blue bg-signal-blue/5 dark:bg-signal-blue/10 shadow-sm"
                      : "border-zinc-200 dark:border-zinc-700 hover:border-zinc-300 dark:hover:border-zinc-600"
                  }`}
                >
                  <input
                    type="radio"
                    {...form.register("neurotype")}
                    value={nt.value}
                    className="sr-only"
                  />
                  <div className="flex flex-col gap-1">
                    <span className="font-semibold text-sm">{nt.label}</span>
                    <span className="text-xs text-zinc-500 dark:text-zinc-400">{nt.desc}</span>
                  </div>
                  {watchNeurotype === nt.value && (
                    <div className="absolute top-2 right-2 h-5 w-5 rounded-full bg-signal-blue flex items-center justify-center">
                      <svg className="h-3 w-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                      </svg>
                    </div>
                  )}
                </label>
              ))}
            </div>
          </section>

          {/* Feature Toggles */}
          <section>
            <div className="flex items-center gap-2 mb-6">
              <Zap className="h-5 w-5 text-signal-blue" />
              <h2 className="text-lg font-semibold">Features</h2>
            </div>

            <div className="space-y-4">
              {/* Focus Mode */}
              <div className="glass rounded-xl p-5">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-violet-100 dark:bg-violet-950">
                      <Zap className="h-5 w-5 text-violet-600 dark:text-violet-400" />
                    </div>
                    <div>
                      <div className="font-medium">Focus Mode</div>
                      <div className="text-xs text-zinc-500">Detects fast channels & generates AI summaries</div>
                    </div>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      {...form.register("focus_mode_enabled")}
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-zinc-200 peer-focus:outline-none rounded-full peer dark:bg-zinc-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-signal-blue" />
                  </label>
                </div>
                {form.watch("focus_mode_enabled") && (
                  <div className="mt-4 pt-4 border-t border-zinc-200 dark:border-zinc-700">
                    <label className="text-sm text-zinc-600 dark:text-zinc-400 block mb-2">
                      Trigger threshold: <strong>{form.watch("focus_threshold")}</strong> messages in 10 min
                    </label>
                    <input
                      type="range"
                      min={10}
                      max={100}
                      step={5}
                      {...form.register("focus_threshold", { valueAsNumber: true })}
                      className="w-full h-2 bg-zinc-200 rounded-lg appearance-none cursor-pointer dark:bg-zinc-700 accent-signal-blue"
                    />
                    <div className="flex justify-between text-xs text-zinc-400 mt-1">
                      <span>10</span>
                      <span>50</span>
                      <span>100</span>
                    </div>
                  </div>
                )}
              </div>

              {/* Social Translator */}
              <div className="glass rounded-xl p-5">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-emerald-100 dark:bg-emerald-950">
                      <Shield className="h-5 w-5 text-emerald-600 dark:text-emerald-400" />
                    </div>
                    <div>
                      <div className="font-medium">Social Translator</div>
                      <div className="text-xs text-zinc-500">Decodes ambiguous messages into plain language</div>
                    </div>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      {...form.register("translator_enabled")}
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-zinc-200 peer-focus:outline-none rounded-full peer dark:bg-zinc-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-signal-blue" />
                  </label>
                </div>
              </div>

              {/* Quiet Hours Digest */}
              <div className="glass rounded-xl p-5">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-rose-100 dark:bg-rose-950">
                      <Bell className="h-5 w-5 text-rose-600 dark:text-rose-400" />
                    </div>
                    <div>
                      <div className="font-medium">Quiet Hours Digest</div>
                      <div className="text-xs text-zinc-500">Batches non-urgent mentions into a daily digest</div>
                    </div>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      {...form.register("digest_enabled")}
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-zinc-200 peer-focus:outline-none rounded-full peer dark:bg-zinc-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-signal-blue" />
                  </label>
                </div>
                {watchDigest && (
                  <div className="mt-4 pt-4 border-t border-zinc-200 dark:border-zinc-700">
                    <label className="text-sm text-zinc-600 dark:text-zinc-400 block mb-2">
                      Delivery time: <strong>{form.watch("digest_hour")}:00</strong>
                    </label>
                    <input
                      type="range"
                      min={0}
                      max={23}
                      step={1}
                      {...form.register("digest_hour", { valueAsNumber: true })}
                      className="w-full h-2 bg-zinc-200 rounded-lg appearance-none cursor-pointer dark:bg-zinc-700 accent-signal-blue"
                    />
                    <div className="flex justify-between text-xs text-zinc-400 mt-1">
                      <span>Midnight</span>
                      <span>Noon</span>
                      <span>11 PM</span>
                    </div>
                  </div>
                )}
              </div>

              {/* Deep Work Protector */}
              <div className="glass rounded-xl p-5">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-amber-100 dark:bg-amber-950">
                      <Moon className="h-5 w-5 text-amber-600 dark:text-amber-400" />
                    </div>
                    <div>
                      <div className="font-medium">Deep Work Protector</div>
                      <div className="text-xs text-zinc-500">Auto-detect focus time from calendar events</div>
                    </div>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      {...form.register("deep_work_auto_detect")}
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-zinc-200 peer-focus:outline-none rounded-full peer dark:bg-zinc-700 peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-signal-blue" />
                  </label>
                </div>
              </div>

              {/* Quiet Hours Time */}
              <div className="glass rounded-xl p-5">
                <div className="flex items-center gap-3 mb-4">
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100 dark:bg-blue-950">
                    <Moon className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                  </div>
                  <div>
                    <div className="font-medium">Quiet Hours</div>
                    <div className="text-xs text-zinc-500">Do not disturb between these hours</div>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-xs text-zinc-500 block mb-1">Start</label>
                    <input
                      type="time"
                      {...form.register("quiet_hours_start")}
                      className="w-full rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 py-2 text-sm"
                    />
                  </div>
                  <div>
                    <label className="text-xs text-zinc-500 block mb-1">End</label>
                    <input
                      type="time"
                      {...form.register("quiet_hours_end")}
                      className="w-full rounded-lg border border-zinc-200 dark:border-zinc-700 bg-white dark:bg-zinc-800 px-3 py-2 text-sm"
                    />
                  </div>
                </div>
              </div>
            </div>
          </section>

          {/* Save */}
          <div className="flex items-center justify-between border-t border-zinc-200 dark:border-zinc-800 pt-8">
            <p className="text-sm text-zinc-500">
              Changes take effect immediately.
            </p>
            <button
              type="submit"
              disabled={saving}
              className="btn-primary flex items-center gap-2"
            >
              <Save className="h-4 w-4" />
              {saving ? "Saving..." : "Save Preferences"}
            </button>
          </div>
        </form>
      </main>
    </div>
  );
}
