import { Toaster } from "sonner";
import { BrowserRouter as Router, Navigate, Route, Routes } from 'react-router-dom';
import AppLayout from './components/dashboard/AppLayout';
import ProtectedRoute from './components/ProtectedRoute';
import Dashboard from './pages/Dashboard';
import SchemaBoard from './pages/SchemaBoard';
import Tables from './pages/Tables';
import Credentials from './pages/Credentials';
import SqlPreview from './pages/SqlPreview';
import Backups from './pages/Backups';
import AppSettings from './pages/AppSettings';
import AuthPage from './pages/AuthPage';
import PageNotFound from "./pages/PageNotFound";
import { AuthProvider } from './lib/AuthContext';
import { ProjectsProvider } from './lib/ProjectsContext';

const DashboardApp = () => (
  <Routes>
    <Route path="/login" element={<AuthPage />} />
    <Route element={<ProtectedRoute />}>
      <Route path="/" element={<Dashboard />} />
      <Route path="/projects/:projectId" element={<AppLayout />}>
        <Route index element={<Navigate to="board" replace />} />
        <Route path="board" element={<SchemaBoard />} />
        <Route path="tables" element={<Tables />} />
        <Route path="credentials" element={<Credentials />} />
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
