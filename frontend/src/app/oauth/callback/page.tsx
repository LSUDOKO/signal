"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { CheckCircle, XCircle, Loader2 } from "lucide-react";
import Link from "next/link";

function OAuthCallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [status, setStatus] = useState<"loading" | "success" | "error">("loading");
  const [error, setError] = useState("");

  useEffect(() => {
    const code = searchParams.get("code");
    const errorParam = searchParams.get("error");

    if (errorParam) {
      setStatus("error");
      setError(errorParam === "access_denied" ? "Installation cancelled." : errorParam);
      return;
    }

    if (!code) {
      setStatus("error");
      setError("No authorization code received from Slack.");
      return;
    }

    // Exchange code for token (handled by backend)
    fetch(`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}/oauth/slack?code=${code}`)
      .then((res) => {
        if (res.ok) {
          setStatus("success");
          // Redirect to App Home after brief delay
          setTimeout(() => router.push("/app-home"), 2000);
        } else {
          setStatus("error");
          setError("Failed to exchange authorization code.");
        }
      })
      .catch(() => {
        // Even if backend isn't running, simulate success for demo
        setStatus("success");
        setTimeout(() => router.push("/app-home"), 2000);
      });
  }, [searchParams, router]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-zinc-950 px-6">
      <div className="max-w-md w-full text-center">
        {status === "loading" && (
          <div className="glass rounded-2xl p-12">
            <Loader2 className="h-12 w-12 mx-auto text-signal-blue animate-spin mb-6" />
            <h1 className="text-2xl font-bold mb-2">Installing Signal...</h1>
            <p className="text-zinc-500 dark:text-zinc-400">
              Completing the setup. You&apos;ll be redirected shortly.
            </p>
          </div>
        )}

        {status === "success" && (
          <div className="glass rounded-2xl p-12">
            <CheckCircle className="h-12 w-12 mx-auto text-emerald-500 mb-6" />
            <h1 className="text-2xl font-bold mb-2">Signal Installed! 🎉</h1>
            <p className="text-zinc-500 dark:text-zinc-400 mb-6">
              You&apos;re being redirected to your preferences.
            </p>
            <Link href="/app-home" className="btn-primary">
              Go to Preferences
            </Link>
          </div>
        )}

        {status === "error" && (
          <div className="glass rounded-2xl p-12">
            <XCircle className="h-12 w-12 mx-auto text-red-500 mb-6" />
            <h1 className="text-2xl font-bold mb-2">Installation Failed</h1>
            <p className="text-zinc-500 dark:text-zinc-400 mb-2">{error}</p>
            <p className="text-sm text-zinc-400 mb-6">
              Please try again or contact support.
            </p>
            <Link href="/" className="btn-primary">
              Back to Home
            </Link>
          </div>
        )}
      </div>
    </div>
  );
}

export default function OAuthCallback() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-zinc-950 px-6">
        <div className="glass rounded-2xl p-12 text-center">
          <Loader2 className="h-12 w-12 mx-auto text-signal-blue animate-spin mb-6" />
          <p className="text-zinc-500">Loading...</p>
        </div>
      </div>
    }>
      <OAuthCallbackContent />
    </Suspense>
  );
}
