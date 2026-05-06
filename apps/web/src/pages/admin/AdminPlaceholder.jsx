export default function AdminPlaceholder({ title }) {
  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      <div className="rounded-3xl border border-border bg-card p-8 shadow-xl shadow-black/10">
        <h1 className="text-2xl font-semibold">{title}</h1>
        <p className="mt-3 text-sm text-muted-foreground">Coming soon.</p>
      </div>
    </div>
  );
}
