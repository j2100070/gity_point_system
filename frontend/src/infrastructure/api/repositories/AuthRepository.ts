import { IAuthRepository } from '@/core/repositories/interfaces';
import { User, LoginRequest, RegisterRequest, AuthResponse } from '@/core/domain/User';
import { axiosInstance, apiClient } from '../client';

export class AuthRepository implements IAuthRepository {
  async login(request: LoginRequest): Promise<AuthResponse> {
    const response = await axiosInstance.post<AuthResponse>('/api/auth/login', request);
    return response.data;
  }

  async register(request: RegisterRequest): Promise<AuthResponse> {
    const response = await axiosInstance.post<AuthResponse>('/api/auth/register', request);
    return response.data;
  }

  async logout(): Promise<void> {
    await axiosInstance.post('/api/auth/logout');
    apiClient.clearCsrfToken();
  }

  async getCurrentUser(): Promise<User> {
    const response = await axiosInstance.get<{ user: User }>('/api/auth/me');
    return response.data.user;
  }
}
