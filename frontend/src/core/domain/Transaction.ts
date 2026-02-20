export interface Transaction {
  id: string;
  from_user_id?: string;
  to_user_id?: string;
  amount: number;
  transaction_type: 'transfer' | 'admin_grant' | 'admin_deduct' | 'system_grant' | 'system_expire' | 'daily_bonus';
  status: 'pending' | 'completed' | 'failed' | 'reversed';
  description?: string;
  created_at: string;
  completed_at?: string;
  from_user?: {
    id: string;
    username: string;
    display_name: string;
  };
  to_user?: {
    id: string;
    username: string;
    display_name: string;
  };
}

export interface TransferRequest {
  to_user_id: string;
  amount: number;
  idempotency_key: string;
  description?: string;
}

export interface TransferResponse {
  message: string;
  transaction: Transaction;
  new_balance: number;
}
