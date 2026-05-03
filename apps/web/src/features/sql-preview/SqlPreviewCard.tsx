import { Button } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";

export function SqlPreviewCard({
  sql,
  onApply,
  isApplying
}: {
  sql: string;
  onApply: () => void;
  isApplying: boolean;
}) {
  return (
    <Card className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-bold text-ink">SQL preview</h2>
          <p className="text-sm text-slate-600">Review the generated PostgreSQL before applying it.</p>
        </div>
        <Button disabled={isApplying} onClick={onApply}>
          {isApplying ? "Applying..." : "Apply schema"}
        </Button>
      </div>
      <div className="rounded-2xl bg-panel p-4 text-sm text-white">
        <pre className="max-h-[320px] overflow-auto whitespace-pre-wrap">{sql || "Generate a preview to inspect SQL."}</pre>
      </div>
    </Card>
  );
}
