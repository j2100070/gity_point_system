export type TransferRequestStatus = 'pending' | 'approved' | 'rejected' | 'cancelled' | 'expired';

export interface TransferRequest {
  id: string;
  from_user_id: string;
  to_user_id: string;
  amount: number;
  message: string;
  status: TransferRequestStatus;
  expires_at: string;
  approved_at?: string;
  rejected_at?: string;
  cancelled_at?: string;
  transaction_id?: string;
  created_at: string;
  updated_at: string;
}

export interface TransferRequestInfo {
  transfer_request: TransferRequest;
  from_user: {
    id: string;
    username: string;
    display_name: string;
    avatar_url?: string;
    balance: number;
  };
  to_user: {
    id: string;
    username: string;
    display_name: string;
    avatar_url?: string;
    balance: number;
  };
}

export interface CreateTransferRequestParams {
  to_user_id: string;
  amount: number;
  message?: string;
  idempotency_key: string;
}

export interface ApproveTransferRequestResponse {
  transfer_request: TransferRequest;
  transaction: {
    id: string;
    from_user_id?: string;
    to_user_id?: string;
    amount: number;
    description: string;
    created_at: string;
  };
  from_user: {
    id: string;
    username: string;
    display_name: string;
    avatar_url?: string;
    balance: number;
  };
  to_user: {
    id: string;
    username: string;
    display_name: string;
    avatar_url?: string;
    balance: number;
  };
}
