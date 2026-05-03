import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useLogin } from "../../api/auth";
import { Button } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import { Input } from "../../components/ui/Input";

export function LoginForm() {
  const navigate = useNavigate();
  const login = useLogin();
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("");

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    login.mutate(
      { username, password },
      {
        onSuccess: () => navigate("/")
      }
    );
  }

  return (
    <Card className="w-full max-w-md p-8">
      <div className="mb-6">
        <p className="text-sm font-semibold uppercase tracking-[0.2em] text-ember">Self-hosted control plane</p>
        <h1 className="mt-2 text-3xl font-black text-ink">Log in to your database launchpad</h1>
      </div>
      <form className="space-y-4" onSubmit={handleSubmit}>
        <label className="block text-sm font-medium text-slate-700">
          Username
          <Input value={username} onChange={(event) => setUsername(event.target.value)} />
        </label>
        <label className="block text-sm font-medium text-slate-700">
          Password
          <Input type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
        </label>
        {login.isError ? <p className="text-sm text-rose-600">{login.error.message}</p> : null}
        <Button className="w-full" disabled={login.isPending} type="submit">
          {login.isPending ? "Signing in..." : "Log in"}
        </Button>
      </form>
    </Card>
  );
}
