import { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Layout } from '@/shared/components/Layout';
import { ProtectedRoute } from '@/shared/components/ProtectedRoute';
import { useAuthStore } from '@/shared/stores/authStore';

// Auth
import { LoginPage } from '@/features/auth/pages/LoginPage';
import { RegisterPage } from '@/features/auth/pages/RegisterPage';

// Dashboard
import { DashboardPage } from '@/features/dashboard/pages/DashboardPage';

// QR Code
import { ReceiveQRPage } from '@/features/qrcode/pages/ReceiveQRPage';
import { ScanQRPage } from '@/features/qrcode/pages/ScanQRPage';

// Points
import { TransferPage } from '@/features/points/pages/TransferPage';
import { HistoryPage } from '@/features/points/pages/HistoryPage';

// Friends
import { FriendsPage } from '@/features/friends/pages/FriendsPage';

// Admin
import { AdminDashboardPage } from '@/features/admin/pages/AdminDashboardPage';
import { AdminUsersPage } from '@/features/admin/pages/AdminUsersPage';
import { AdminTransactionsPage } from '@/features/admin/pages/AdminTransactionsPage';

function App() {
  const { loadUser, isAuthenticated } = useAuthStore();

  useEffect(() => {
    // アプリ起動時にログイン状態を確認
    loadUser();
  }, [loadUser]);

  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          {/* Public routes */}
          <Route
            path="/login"
            element={isAuthenticated ? <Navigate to="/dashboard" replace /> : <LoginPage />}
          />
          <Route
            path="/register"
            element={isAuthenticated ? <Navigate to="/dashboard" replace /> : <RegisterPage />}
          />

          {/* Protected routes */}
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <DashboardPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/qr/receive"
            element={
              <ProtectedRoute>
                <ReceiveQRPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/qr/send"
            element={
              <ProtectedRoute>
                <ReceiveQRPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/qr/scan"
            element={
              <ProtectedRoute>
                <ScanQRPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/transfer"
            element={
              <ProtectedRoute>
                <TransferPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/history"
            element={
              <ProtectedRoute>
                <HistoryPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/friends"
            element={
              <ProtectedRoute>
                <FriendsPage />
              </ProtectedRoute>
            }
          />

          {/* Admin routes */}
          <Route
            path="/admin"
            element={
              <ProtectedRoute adminOnly>
                <AdminDashboardPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/admin/users"
            element={
              <ProtectedRoute adminOnly>
                <AdminUsersPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/admin/transactions"
            element={
              <ProtectedRoute adminOnly>
                <AdminTransactionsPage />
              </ProtectedRoute>
            }
          />

          {/* Default redirect */}
          <Route
            path="/"
            element={<Navigate to={isAuthenticated ? '/dashboard' : '/login'} replace />}
          />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  );
}

export default App;
