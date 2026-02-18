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

  async getAllUsers(offset: number = 0, limit: number = 20, search?: string, sortBy?: string, sortOrder?: string): Promise<{ users: User[]; total: number }> {
    const params = new URLSearchParams();
    params.set('offset', String(offset));
    params.set('limit', String(limit));
    if (search) params.set('search', search);
    if (sortBy) params.set('sort_by', sortBy);
    if (sortOrder) params.set('sort_order', sortOrder);
    const response = await axiosInstance.get<{ users: User[]; total: number }>(
      `/api/admin/users?${params.toString()}`
    );
    return response.data;
  }

  async getAllTransactions(
    offset: number = 0,
    limit: number = 20,
    transactionType?: string,
    dateFrom?: string,
    dateTo?: string,
    sortBy?: string,
    sortOrder?: string,
  ): Promise<{ transactions: Transaction[]; total: number }> {
    const params = new URLSearchParams();
    params.set('offset', String(offset));
    params.set('limit', String(limit));
    if (transactionType) params.set('transaction_type', transactionType);
    if (dateFrom) params.set('date_from', dateFrom);
    if (dateTo) params.set('date_to', dateTo);
    if (sortBy) params.set('sort_by', sortBy);
    if (sortOrder) params.set('sort_order', sortOrder);
    const response = await axiosInstance.get<{ transactions: Transaction[]; total: number }>(
      `/api/admin/transactions?${params.toString()}`
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
