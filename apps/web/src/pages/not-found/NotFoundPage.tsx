import { Link } from "react-router-dom";

export function NotFoundPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4">
      <h1 className="text-3xl font-black text-ink">Page not found</h1>
      <Link className="text-ember" to="/">
        Back home
      </Link>
    </div>
  );
}
