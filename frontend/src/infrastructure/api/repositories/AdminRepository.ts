import { IAdminRepository } from '@/core/repositories/interfaces';
import { User } from '@/core/domain/User';
import { Transaction } from '@/core/domain/Transaction';
import { axiosInstance } from '../client';

export class AdminRepository implements IAdminRepository {
  async grantPoints(target_user_id: string, amount: number, description?: string): Promise<any> {
    const response = await axiosInstance.post('/api/admin/points/grant', {
      user_id: target_user_id,
      amount,
      description: description || 'Admin point grant',
      idempotency_key: `grant-${Date.now()}-${Math.random().toString(36).substring(7)}`,
    });
    return response.data;
  }

  async deductPoints(target_user_id: string, amount: number, description?: string): Promise<any> {
    const response = await axiosInstance.post('/api/admin/points/deduct', {
      user_id: target_user_id,
      amount,
      description: description || 'Admin point deduction',
      idempotency_key: `deduct-${Date.now()}-${Math.random().toString(36).substring(7)}`,
    });
    return response.data;
  }

  async getAllUsers(offset: number = 0, limit: number = 20): Promise<{ users: User[]; total: number }> {
    const response = await axiosInstance.get<{ users: User[]; total: number }>(
      `/api/admin/users?offset=${offset}&limit=${limit}`
    );
    return response.data;
  }

  async getAllTransactions(offset: number = 0, limit: number = 20): Promise<{ transactions: Transaction[]; total: number }> {
    const response = await axiosInstance.get<{ transactions: Transaction[]; total: number }>(
      `/api/admin/transactions?offset=${offset}&limit=${limit}`
    );
    return response.data;
  }

  async changeUserRole(target_user_id: string, new_role: string): Promise<any> {
    const response = await axiosInstance.put(`/api/admin/users/${target_user_id}/role`, {
      role: new_role,
    });
    return response.data;
  }

  async deactivateUser(target_user_id: string): Promise<any> {
    const response = await axiosInstance.post(`/api/admin/users/${target_user_id}/deactivate`);
    return response.data;
  }

  async getAnalytics(days: number = 30): Promise<any> {
    const response = await axiosInstance.get(`/api/admin/analytics?days=${days}`);
    return response.data;
  }
}
