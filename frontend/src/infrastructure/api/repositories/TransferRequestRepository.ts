import {
  TransferRequest,
  TransferRequestInfo,
  CreateTransferRequestParams,
  ApproveTransferRequestResponse,
} from '@/core/domain/TransferRequest';
import { axiosInstance } from '../client';

export class TransferRequestRepository {
  // 個人QRコードを取得
  async getPersonalQRCode(): Promise<{ qr_code: string; display_name: string; username: string }> {
    const response = await axiosInstance.get('/api/transfer-requests/personal-qr');
    return response.data;
  }

  // 送金リクエストを作成
  async createTransferRequest(params: CreateTransferRequestParams): Promise<TransferRequestInfo> {
    const response = await axiosInstance.post('/api/transfer-requests', params);
    return response.data;
  }

  // 送金リクエストを承認
  async approveTransferRequest(requestId: string): Promise<ApproveTransferRequestResponse> {
    const response = await axiosInstance.post(`/api/transfer-requests/${requestId}/approve`);
    return response.data;
  }

  // 送金リクエストを拒否
  async rejectTransferRequest(requestId: string): Promise<{ transfer_request: TransferRequest }> {
    const response = await axiosInstance.post(`/api/transfer-requests/${requestId}/reject`);
    return response.data;
  }

  // 送金リクエストをキャンセル
  async cancelTransferRequest(requestId: string): Promise<{ transfer_request: TransferRequest }> {
    const response = await axiosInstance.delete(`/api/transfer-requests/${requestId}`);
    return response.data;
  }

  // 受取人宛の承認待ちリクエスト一覧を取得
  async getPendingRequests(offset: number = 0, limit: number = 20): Promise<{ requests: TransferRequestInfo[] }> {
    const response = await axiosInstance.get(
      `/api/transfer-requests/pending?offset=${offset}&limit=${limit}`
    );
    return response.data;
  }

  // 送信者が送ったリクエスト一覧を取得
  async getSentRequests(offset: number = 0, limit: number = 20): Promise<{ requests: TransferRequestInfo[] }> {
    const response = await axiosInstance.get(
      `/api/transfer-requests/sent?offset=${offset}&limit=${limit}`
    );
    return response.data;
  }

  // 送金リクエスト詳細を取得
  async getRequestDetail(requestId: string): Promise<TransferRequestInfo> {
    const response = await axiosInstance.get(`/api/transfer-requests/${requestId}`);
    return response.data;
  }

  // 承認待ちリクエスト数を取得
  async getPendingRequestCount(): Promise<{ count: number }> {
    const response = await axiosInstance.get('/api/transfer-requests/pending/count');
    return response.data;
  }
}
