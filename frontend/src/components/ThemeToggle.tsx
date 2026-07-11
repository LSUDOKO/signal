"use client";

import { useTheme } from "next-themes";
import { Sun, Moon } from "lucide-react";
import { useEffect, useState } from "react";

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return (
      <div className="h-9 w-9 rounded-lg border border-zinc-200 dark:border-zinc-700" />
    );
  }

  return (
    <button
      onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
      className="flex h-9 w-9 items-center justify-center rounded-lg border border-zinc-200 dark:border-zinc-700 hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-all"
      aria-label="Toggle theme"
    >
      {theme === "dark" ? (
        <Sun className="h-4 w-4 text-zinc-400" />
      ) : (
        <Moon className="h-4 w-4 text-zinc-600" />
      )}
    </button>
  );
}
