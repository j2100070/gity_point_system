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

// Transfer Requests
import { PersonalQRPage } from '@/features/transfer-requests/pages/PersonalQRPage';
import { TransferRequestsPage } from '@/features/transfer-requests/pages/TransferRequestsPage';

// Points
import { TransferPage } from '@/features/points/pages/TransferPage';
import { HistoryPage } from '@/features/points/pages/HistoryPage';

// Friends
import { FriendsPage } from '@/features/friends/pages/FriendsPage';

// Settings
import { SettingsPage } from '@/features/settings/pages/SettingsPage';
import { VerifyEmailPage } from '@/features/settings/pages/VerifyEmailPage';

// Products
import { ProductsPage } from '@/features/products/pages/ProductsPage';
import { ExchangePage } from '@/features/products/pages/ExchangePage';
import { ExchangeHistoryPage } from '@/features/products/pages/ExchangeHistoryPage';

// Admin
import { AdminDashboardPage } from '@/features/admin/pages/AdminDashboardPage';
import { AdminUsersPage } from '@/features/admin/pages/AdminUsersPage';
import { AdminTransactionsPage } from '@/features/admin/pages/AdminTransactionsPage';
import { AdminCategoriesPage } from '@/features/admin/pages/AdminCategoriesPage';
import { AdminProductsPage } from '@/features/products/pages/AdminProductsPage';

function App() {
  const { loadUser, isAuthenticated, isLoading } = useAuthStore();

  useEffect(() => {
    // アプリ起動時にログイン状態を確認
    loadUser();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 認証状態の確認中はスピナーを表示
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    );
  }
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
            path="/qr/personal"
            element={
              <ProtectedRoute>
                <PersonalQRPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/transfer-requests"
            element={
              <ProtectedRoute>
                <TransferRequestsPage />
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

          {/* Settings routes */}
          <Route
            path="/settings"
            element={
              <ProtectedRoute>
                <SettingsPage />
              </ProtectedRoute>
            }
          />
          <Route path="/verify-email" element={<VerifyEmailPage />} />

          {/* Product routes */}
          <Route
            path="/products"
            element={
              <ProtectedRoute>
                <ProductsPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/products/:productId/exchange"
            element={
              <ProtectedRoute>
                <ExchangePage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/products/exchanges"
            element={
              <ProtectedRoute>
                <ExchangeHistoryPage />
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
          <Route
            path="/admin/products"
            element={
              <ProtectedRoute adminOnly>
                <AdminProductsPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/admin/categories"
            element={
              <ProtectedRoute adminOnly>
                <AdminCategoriesPage />
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
