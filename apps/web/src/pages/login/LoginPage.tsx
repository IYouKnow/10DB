import { LoginForm } from "../../features/auth/LoginForm";

export function LoginPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-[radial-gradient(circle_at_top,_rgba(141,210,200,0.45),_transparent_35%),linear-gradient(180deg,_#08111f_0%,_#18263d_100%)] px-6">
      <LoginForm />
    </div>
  );
}
