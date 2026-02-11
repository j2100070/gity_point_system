import { create } from 'zustand';
import { User } from '@/core/domain/User';
import { AuthRepository } from '@/infrastructure/api/repositories/AuthRepository';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (username: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string, displayName: string) => Promise<void>;
  logout: () => Promise<void>;
  loadUser: () => Promise<void>;
  clearError: () => void;
}

const authRepository = new AuthRepository();

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  isLoading: true, // 初期状態はローディング中（loadUser完了まで）
  error: null,

  login: async (username: string, password: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await authRepository.login({ username, password });
      set({
        user: response.user,
        isAuthenticated: true,
        isLoading: false,
      });
    } catch (error: any) {
      set({
        error: error.response?.data?.error || 'ログインに失敗しました',
        isLoading: false,
      });
      throw error;
    }
  },

  register: async (username: string, email: string, password: string, displayName: string) => {
    set({ isLoading: true, error: null });
    try {
      const response = await authRepository.register({
        username,
        email,
        password,
        display_name: displayName,
      });
      set({
        user: response.user,
        isAuthenticated: true,
        isLoading: false,
      });
    } catch (error: any) {
      set({
        error: error.response?.data?.error || '登録に失敗しました',
        isLoading: false,
      });
      throw error;
    }
  },

  logout: async () => {
    try {
      await authRepository.logout();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      set({
        user: null,
        isAuthenticated: false,
      });
    }
  },

  loadUser: async () => {
    set({ isLoading: true });
    try {
      const user = await authRepository.getCurrentUser();
      set({
        user,
        isAuthenticated: true,
        isLoading: false,
      });
    } catch (error) {
      set({
        user: null,
        isAuthenticated: false,
        isLoading: false,
      });
    }
  },

  clearError: () => set({ error: null }),
}));
