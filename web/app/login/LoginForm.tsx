"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";

type Step = "email" | "code";

export function LoginForm() {
  const router = useRouter();
  const [step, setStep] = useState<Step>("email");
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [pending, startTransition] = useTransition();

  async function sendOTP(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    startTransition(async () => {
      const res = await fetch("/api/auth/otp?action=send", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });
      if (res.status === 204) {
        setStep("code");
      } else {
        setError("Failed to send code. Check the email address.");
      }
    });
  }

  async function verifyOTP(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    startTransition(async () => {
      const res = await fetch("/api/auth/otp?action=verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, code }),
      });
      if (res.ok) {
        router.push("/account");
      } else {
        setError("Invalid or expired code. Try again.");
      }
    });
  }

  if (step === "email") {
    return (
      <form onSubmit={sendOTP} className="flex flex-col gap-4">
        <div>
          <label className="mb-1.5 block text-xs text-[var(--ink-400)]" htmlFor="email">
            Email address
          </label>
          <input
            id="email"
            type="email"
            required
            autoFocus
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@example.com"
            className="w-full rounded-md border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-3 py-2.5 text-sm text-[var(--ink-100)] placeholder:text-[var(--ink-600)] focus:border-[var(--ink-400)] focus:outline-none"
          />
        </div>
        {error && <p className="text-xs text-[var(--signal-down)]">{error}</p>}
        <button
          type="submit"
          disabled={pending}
          className="rounded-md bg-[var(--ink-200)] px-4 py-2.5 text-sm font-medium text-[var(--canvas-base)] hover:bg-[var(--ink-100)] disabled:opacity-50 transition-colors"
        >
          {pending ? "Sending…" : "Send code"}
        </button>
      </form>
    );
  }

  return (
    <form onSubmit={verifyOTP} className="flex flex-col gap-4">
      <p className="text-xs text-[var(--ink-400)]">
        Enter the 6-digit code sent to <span className="text-[var(--ink-200)]">{email}</span>.
      </p>
      <input
        type="text"
        inputMode="numeric"
        pattern="[0-9]{6}"
        maxLength={6}
        required
        autoFocus
        value={code}
        onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))}
        placeholder="000000"
        className="w-full rounded-md border border-[var(--ink-600)] bg-[var(--canvas-raised)] px-3 py-2.5 text-center font-mono text-2xl tracking-[0.5em] text-[var(--ink-100)] placeholder:text-[var(--ink-700)] focus:border-[var(--ink-400)] focus:outline-none"
      />
      {error && <p className="text-xs text-[var(--signal-down)]">{error}</p>}
      <button
        type="submit"
        disabled={pending || code.length !== 6}
        className="rounded-md bg-[var(--ink-200)] px-4 py-2.5 text-sm font-medium text-[var(--canvas-base)] hover:bg-[var(--ink-100)] disabled:opacity-50 transition-colors"
      >
        {pending ? "Verifying…" : "Sign in"}
      </button>
      <button
        type="button"
        onClick={() => { setStep("email"); setCode(""); setError(""); }}
        className="text-xs text-[var(--ink-500)] hover:text-[var(--ink-300)] transition-colors"
      >
        Use a different email
      </button>
    </form>
  );
}
