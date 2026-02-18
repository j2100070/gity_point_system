import { IQRCodeRepository } from '@/core/repositories/interfaces';
import { QRCode, GenerateQRRequest, GenerateQRResponse, ScanQRRequest, ScanQRResponse } from '@/core/domain/QRCode';
import { axiosInstance } from '../client';

export class QRCodeRepository implements IQRCodeRepository {
  async generateReceiveQR(request: GenerateQRRequest): Promise<GenerateQRResponse> {
    const response = await axiosInstance.post<GenerateQRResponse>('/api/qrcodes/receive', request);
    return response.data;
  }

  async generateSendQR(request: GenerateQRRequest): Promise<GenerateQRResponse> {
    const response = await axiosInstance.post<GenerateQRResponse>('/api/qrcodes/send', request);
    return response.data;
  }

  async scanQR(request: ScanQRRequest): Promise<ScanQRResponse> {
    const response = await axiosInstance.post<ScanQRResponse>('/api/qrcodes/scan', request);
    return response.data;
  }

  async getHistory(offset: number = 0, limit: number = 20): Promise<{ qr_codes: QRCode[] }> {
    const response = await axiosInstance.get<{ qr_codes: QRCode[] }>(
      `/api/qrcodes/history?offset=${offset}&limit=${limit}`
    );
    return response.data;
  }
}
