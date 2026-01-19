export interface QRCode {
  id: string;
  user_id: string;
  code: string;
  qr_data: string;
  amount?: number;
  qr_type: 'receive' | 'send';
  expires_at: string;
  used_at?: string;
  used_by_user_id?: string;
  created_at: string;
}

export interface GenerateQRRequest {
  amount?: number;
}

export interface GenerateQRResponse {
  qr_code: QRCode;
}

export interface ScanQRRequest {
  qr_code: string;
  amount?: number;
  idempotency_key: string;
}

export interface ScanQRResponse {
  message: string;
  transaction: {
    id: string;
    amount: number;
    status: string;
  };
}
