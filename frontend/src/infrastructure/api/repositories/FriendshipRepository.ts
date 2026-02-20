import { IFriendshipRepository } from '@/core/repositories/interfaces';
import { FriendInfo, PendingRequestInfo, FriendRequestRequest, FriendActionRequest, Friendship } from '@/core/domain/Friendship';
import { axiosInstance } from '../client';

export interface UserBasicInfo {
  id: string;
  username: string;
  display_name: string;
  avatar_url?: string;
  avatar_type?: string;
}

export class FriendshipRepository implements IFriendshipRepository {
  async sendRequest(request: FriendRequestRequest): Promise<{ message: string; friendship: Friendship }> {
    const response = await axiosInstance.post<{ message: string; friendship: Friendship }>(
      '/api/friends/requests',
      request
    );
    return response.data;
  }

  async acceptRequest(request: FriendActionRequest): Promise<{ message: string }> {
    const response = await axiosInstance.post<{ message: string }>(
      `/api/friends/requests/${request.friendship_id}/accept`
    );
    return response.data;
  }

  async rejectRequest(request: FriendActionRequest): Promise<{ message: string }> {
    const response = await axiosInstance.post<{ message: string }>(
      `/api/friends/requests/${request.friendship_id}/reject`
    );
    return response.data;
  }

  async getFriends(offset: number = 0, limit: number = 20): Promise<{ friends: FriendInfo[] }> {
    const response = await axiosInstance.get<{ friends: FriendInfo[] }>(
      `/api/friends?offset=${offset}&limit=${limit}`
    );
    return response.data;
  }

  async getPendingRequests(offset: number = 0, limit: number = 20): Promise<{ requests: PendingRequestInfo[] }> {
    const response = await axiosInstance.get<{ requests: PendingRequestInfo[] }>(
      `/api/friends/requests?offset=${offset}&limit=${limit}`
    );
    return response.data;
  }

  async removeFriend(request: FriendActionRequest): Promise<{ message: string }> {
    const response = await axiosInstance.delete<{ message: string }>(
      `/api/friends/${request.friendship_id}`
    );
    return response.data;
  }

  async getUserById(userId: string): Promise<{ user: UserBasicInfo }> {
    const response = await axiosInstance.get<{ user: UserBasicInfo }>(
      `/api/users/${userId}`
    );
    return response.data;
  }

  async getPendingRequestCount(): Promise<{ count: number }> {
    const response = await axiosInstance.get<{ count: number }>(
      '/api/friends/requests/count'
    );
    return response.data;
  }
}
