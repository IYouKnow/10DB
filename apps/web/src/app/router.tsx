import { Navigate, Outlet, Route, Routes, useLocation } from "react-router-dom";
import { useMe } from "../api/auth";
import { DashboardPage } from "../pages/dashboard/DashboardPage";
import { LoginPage } from "../pages/login/LoginPage";
import { NotFoundPage } from "../pages/not-found/NotFoundPage";
import { ProjectDetailPage } from "../pages/project-detail/ProjectDetailPage";

function ProtectedLayout() {
  const location = useLocation();
  const me = useMe();

  if (me.isLoading) {
    return <div className="flex min-h-screen items-center justify-center text-slate-600">Loading dashboard...</div>;
  }

  if (me.isError) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  return <Outlet />;
}

export function AppRouter() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<ProtectedLayout />}>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/projects/:projectId" element={<ProjectDetailPage />} />
      </Route>
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  );
}
