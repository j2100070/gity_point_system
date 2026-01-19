import { IFriendshipRepository } from '@/core/repositories/interfaces';
import { Friendship, FriendRequestRequest, FriendActionRequest } from '@/core/domain/Friendship';
import { axiosInstance } from '../client';

export class FriendshipRepository implements IFriendshipRepository {
  async sendRequest(request: FriendRequestRequest): Promise<{ message: string; friendship: Friendship }> {
    const response = await axiosInstance.post<{ message: string; friendship: Friendship }>(
      '/api/friends/request',
      request
    );
    return response.data;
  }

  async acceptRequest(request: FriendActionRequest): Promise<{ message: string }> {
    const response = await axiosInstance.post<{ message: string }>('/api/friends/accept', request);
    return response.data;
  }

  async rejectRequest(request: FriendActionRequest): Promise<{ message: string }> {
    const response = await axiosInstance.post<{ message: string }>('/api/friends/reject', request);
    return response.data;
  }

  async getFriends(offset: number = 0, limit: number = 20): Promise<{ friends: Friendship[] }> {
    const response = await axiosInstance.get<{ friends: Friendship[] }>(
      `/api/friends?offset=${offset}&limit=${limit}`
    );
    return response.data;
  }

  async getPendingRequests(offset: number = 0, limit: number = 20): Promise<{ requests: Friendship[] }> {
    const response = await axiosInstance.get<{ requests: Friendship[] }>(
      `/api/friends/pending?offset=${offset}&limit=${limit}`
    );
    return response.data;
  }

  async removeFriend(request: FriendActionRequest): Promise<{ message: string }> {
    const response = await axiosInstance.delete<{ message: string }>('/api/friends/remove', { data: request });
    return response.data;
  }
}
