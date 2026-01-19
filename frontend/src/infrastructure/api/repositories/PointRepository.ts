import { IPointRepository } from '@/core/repositories/interfaces';
import { Transaction, TransferRequest, TransferResponse } from '@/core/domain/Transaction';
import { User } from '@/core/domain/User';
import { axiosInstance } from '../client';

export class PointRepository implements IPointRepository {
  async getBalance(): Promise<{ balance: number; user: User }> {
    const response = await axiosInstance.get<{ balance: number; user: User }>('/api/points/balance');
    return response.data;
  }

  async transfer(request: TransferRequest): Promise<TransferResponse> {
    const response = await axiosInstance.post<TransferResponse>('/api/points/transfer', request);
    return response.data;
  }

  async getHistory(offset: number = 0, limit: number = 20): Promise<{ transactions: Transaction[]; total: number }> {
    const response = await axiosInstance.get<{ transactions: Transaction[]; total: number }>(
      `/api/points/history?offset=${offset}&limit=${limit}`
    );
    return response.data;
  }
}
