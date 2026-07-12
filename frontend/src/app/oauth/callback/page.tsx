"use client";

import { Suspense, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { Loader2 } from "lucide-react";
import Link from "next/link";

function OAuthCallbackContent() {
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

    // Redirect browser to the API's OAuth handler for token exchange
    // The API will exchange the code and redirect to /app-home
    window.location.href = `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}/oauth/slack?code=${code}`;
  }, [searchParams]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-zinc-950 px-6">
      <div className="max-w-md w-full text-center">
        <div className="glass rounded-2xl p-12">
          <Loader2 className="h-12 w-12 mx-auto text-signal-blue animate-spin mb-6" />
          <h1 className="text-2xl font-bold mb-2">Installing Signal...</h1>
          <p className="text-zinc-500 dark:text-zinc-400">
            Completing the setup. You&apos;ll be redirected shortly.
          </p>
          {status === "error" && (
            <>
              <p className="text-red-500 mt-4 text-sm">{error}</p>
              <Link href="/" className="btn-primary mt-6 inline-block">Back to Home</Link>
            </>
          )}
        </div>
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
