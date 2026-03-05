import { lazy, Suspense } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { ShockUIProvider } from "@/providers/ShockUIProvider";
import { ErrorBoundary } from "@/components/ErrorBoundary";
import { ProtectedRoute, AdminRoute } from "@/components/auth/ProtectedRoute";
import { AppShell } from "@/components/layout/AppShell";
import { LoginPage } from "@/pages/LoginPage";
import { Skeleton } from "@/components/ui/skeleton";

// Lazy-load pages for code splitting
const NodesPage = lazy(() => import("@/pages/NodesPage").then((m) => ({ default: m.NodesPage })));
const NodeDetailPage = lazy(() => import("@/pages/NodeDetailPage").then((m) => ({ default: m.NodeDetailPage })));
const UploadPage = lazy(() => import("@/pages/UploadPage").then((m) => ({ default: m.UploadPage })));
const AdminDashboardPage = lazy(() => import("@/pages/AdminDashboardPage").then((m) => ({ default: m.AdminDashboardPage })));
const AdminLocationsPage = lazy(() => import("@/pages/AdminLocationsPage").then((m) => ({ default: m.AdminLocationsPage })));
const AdminLockerPage = lazy(() => import("@/pages/AdminLockerPage").then((m) => ({ default: m.AdminLockerPage })));
const AdminTracePage = lazy(() => import("@/pages/AdminTracePage").then((m) => ({ default: m.AdminTracePage })));
const SettingsPage = lazy(() => import("@/pages/SettingsPage").then((m) => ({ default: m.SettingsPage })));

function PageLoader() {
  return (
    <div className="space-y-4 p-6">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-64" />
    </div>
  );
}

export function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <ShockUIProvider>
          <Routes>
            <Route path="/ui/login" element={<LoginPage />} />

            <Route
              element={
                <ProtectedRoute>
                  <AppShell />
                </ProtectedRoute>
              }
            >
              <Route
                path="/ui/nodes"
                element={
                  <Suspense fallback={<PageLoader />}>
                    <NodesPage />
                  </Suspense>
                }
              />
              <Route
                path="/ui/nodes/:id"
                element={
                  <Suspense fallback={<PageLoader />}>
                    <NodeDetailPage />
                  </Suspense>
                }
              />
              <Route
                path="/ui/upload"
                element={
                  <Suspense fallback={<PageLoader />}>
                    <UploadPage />
                  </Suspense>
                }
              />
              <Route
                path="/ui/settings"
                element={
                  <Suspense fallback={<PageLoader />}>
                    <SettingsPage />
                  </Suspense>
                }
              />

              {/* Admin routes */}
              <Route
                path="/ui/admin"
                element={
                  <AdminRoute>
                    <Suspense fallback={<PageLoader />}>
                      <AdminDashboardPage />
                    </Suspense>
                  </AdminRoute>
                }
              />
              <Route
                path="/ui/admin/locations"
                element={
                  <AdminRoute>
                    <Suspense fallback={<PageLoader />}>
                      <AdminLocationsPage />
                    </Suspense>
                  </AdminRoute>
                }
              />
              <Route
                path="/ui/admin/locker"
                element={
                  <AdminRoute>
                    <Suspense fallback={<PageLoader />}>
                      <AdminLockerPage />
                    </Suspense>
                  </AdminRoute>
                }
              />
              <Route
                path="/ui/admin/trace"
                element={
                  <AdminRoute>
                    <Suspense fallback={<PageLoader />}>
                      <AdminTracePage />
                    </Suspense>
                  </AdminRoute>
                }
              />
            </Route>

            {/* Redirects */}
            <Route path="/ui" element={<Navigate to="/ui/nodes" replace />} />
            <Route path="/ui/" element={<Navigate to="/ui/nodes" replace />} />
            <Route path="*" element={<Navigate to="/ui/nodes" replace />} />
          </Routes>
        </ShockUIProvider>
      </BrowserRouter>
    </ErrorBoundary>
  );
}
