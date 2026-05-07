import { Toaster } from "sonner";
import { BrowserRouter as Router, Navigate, Route, Routes } from 'react-router-dom';
import AdminRoute from './components/AdminRoute';
import AdminLayout from './components/admin/AdminLayout';
import AppLayout from './components/dashboard/AppLayout';
import ProtectedRoute from './components/ProtectedRoute';
import Dashboard from './pages/Dashboard';
import SchemaBoard from './pages/SchemaBoard';
import DatabaseSchemaBoard from './pages/DatabaseSchemaBoard';
import Tables from './pages/Tables';
import Credentials from './pages/Credentials';
import SqlPreview from './pages/SqlPreview';
import Backups from './pages/Backups';
import AppSettings from './pages/AppSettings';
import AuthPage from './pages/AuthPage';
import PageNotFound from "./pages/PageNotFound";
import AdminOverview from './pages/admin/AdminOverview';
import AdminPlaceholder from './pages/admin/AdminPlaceholder';
import AdminServers from './pages/admin/AdminServers';
import { AuthProvider } from './lib/AuthContext';
import { ProjectsProvider } from './lib/ProjectsContext';

const DashboardApp = () => (
  <Routes>
    <Route path="/login" element={<AuthPage />} />
    <Route element={<ProtectedRoute />}>
      <Route path="/" element={<Dashboard />} />
      <Route element={<AdminRoute />}>
        <Route path="/admin" element={<AdminLayout />}>
          <Route index element={<AdminOverview />} />
          <Route path="servers" element={<AdminServers />} />
          <Route path="users" element={<AdminPlaceholder title="Users" />} />
          <Route path="projects" element={<AdminPlaceholder title="Projects" />} />
          <Route path="databases" element={<AdminPlaceholder title="Databases" />} />
        </Route>
      </Route>
      <Route path="/projects/:projectId" element={<AppLayout />}>
        <Route index element={<Navigate to="board" replace />} />
        <Route path="board" element={<SchemaBoard />} />
        <Route path="databases/:databaseId/schema" element={<DatabaseSchemaBoard />} />
        <Route path="databases/:databaseId/credentials" element={<Credentials />} />
        <Route path="tables" element={<Tables />} />
        <Route path="sql" element={<SqlPreview />} />
        <Route path="backups" element={<Backups />} />
        <Route path="settings" element={<AppSettings />} />
      </Route>
    </Route>
    <Route path="*" element={<PageNotFound />} />
  </Routes>
);

function App() {
  return (
    <AuthProvider>
      <ProjectsProvider>
        <Router>
          <DashboardApp />
          <Toaster theme="dark" position="bottom-right" />
        </Router>
      </ProjectsProvider>
    </AuthProvider>
  );
}

export default App;
