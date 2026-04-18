import type { ModelSummary } from "@/lib/api";

const MODEL_TYPE_LABEL: Record<string, string> = {
  chat:      "Chat",
  embedding: "Embedding",
  image:     "Image",
};

export function ModelList({ models }: { models: ModelSummary[] }) {
  const active = models.filter((m) => m.active);

  if (active.length === 0) {
    return (
      <p className="text-sm text-[var(--ink-400)]">No active models configured.</p>
    );
  }

  return (
    <div className="overflow-hidden rounded-lg border border-[var(--ink-600)]">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-[var(--ink-600)] bg-[var(--canvas-sunken)]">
            <th className="px-4 py-2 text-left font-medium text-[var(--ink-300)]">Model</th>
            <th className="px-4 py-2 text-left font-medium text-[var(--ink-300)]">Type</th>
          </tr>
        </thead>
        <tbody>
          {active.map((m, idx) => (
            <tr
              key={m.model_id}
              className={`border-b border-[var(--ink-600)] last:border-0 ${
                idx % 2 === 0 ? "bg-[var(--canvas-raised)]" : "bg-[var(--canvas-base)]"
              }`}
            >
              <td className="px-4 py-2 font-mono text-xs text-[var(--ink-200)]">
                {m.model_id}
              </td>
              <td className="px-4 py-2 text-[var(--ink-300)]">
                {MODEL_TYPE_LABEL[m.model_type] ?? m.model_type}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
