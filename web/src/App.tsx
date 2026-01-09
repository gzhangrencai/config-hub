import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import MainLayout from './layouts/MainLayout'
import ProjectListPage from './pages/ProjectListPage'
import ProjectDetailPage from './pages/ProjectDetailPage'
import ConfigEditorPage from './pages/ConfigEditorPage'
import VersionHistoryPage from './pages/VersionHistoryPage'
import ProjectSettingsPage from './pages/ProjectSettingsPage'
import KeyManagementPage from './pages/KeyManagementPage'
import AuditLogPage from './pages/AuditLogPage'
import ReleasePage from './pages/ReleasePage'
import LoginPage from './pages/LoginPage'
import { useAuthStore } from './stores/auth'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuthStore()
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" />
}

function App() {
  // GitHub Pages 部署在子路径下，需要设置 basename
  const basename = import.meta.env.BASE_URL || '/'
  
  return (
    <BrowserRouter basename={basename}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route
          path="/"
          element={
            <PrivateRoute>
              <MainLayout />
            </PrivateRoute>
          }
        >
          <Route index element={<Navigate to="/projects" replace />} />
          <Route path="projects" element={<ProjectListPage />} />
          <Route path="projects/:projectId" element={<ProjectDetailPage />} />
          <Route path="projects/:projectId/settings" element={<ProjectSettingsPage />} />
          <Route path="projects/:projectId/keys" element={<KeyManagementPage />} />
          <Route path="projects/:projectId/audit" element={<AuditLogPage />} />
          <Route path="configs/:configId" element={<ConfigEditorPage />} />
          <Route path="configs/:configId/versions" element={<VersionHistoryPage />} />
          <Route path="configs/:configId/release" element={<ReleasePage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
