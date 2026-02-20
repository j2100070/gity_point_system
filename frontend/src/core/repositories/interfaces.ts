import { User, LoginRequest, RegisterRequest, AuthResponse } from '../domain/User';
import { Transaction, TransferRequest, TransferResponse } from '../domain/Transaction';
import { QRCode, GenerateQRRequest, GenerateQRResponse, ScanQRRequest, ScanQRResponse } from '../domain/QRCode';
import { Friendship, FriendInfo, PendingRequestInfo, FriendRequestRequest, FriendActionRequest } from '../domain/Friendship';

export interface IAuthRepository {
  login(request: LoginRequest): Promise<AuthResponse>;
  register(request: RegisterRequest): Promise<AuthResponse>;
  logout(): Promise<void>;
  getCurrentUser(): Promise<User>;
}

export interface IPointRepository {
  getBalance(): Promise<{ balance: number; user: User }>;
  transfer(request: TransferRequest): Promise<TransferResponse>;
  getHistory(offset?: number, limit?: number): Promise<{ transactions: Transaction[]; total: number }>;
  getExpiringPoints(): Promise<{ expiring_points: { amount: number; expires_at: string }[]; total_expiring: number }>;
}

export interface IQRCodeRepository {
  generateReceiveQR(request: GenerateQRRequest): Promise<GenerateQRResponse>;
  generateSendQR(request: GenerateQRRequest): Promise<GenerateQRResponse>;
  scanQR(request: ScanQRRequest): Promise<ScanQRResponse>;
  getHistory(offset?: number, limit?: number): Promise<{ qr_codes: QRCode[] }>;
}

export interface IFriendshipRepository {
  sendRequest(request: FriendRequestRequest): Promise<{ message: string; friendship: Friendship }>;
  acceptRequest(request: FriendActionRequest): Promise<{ message: string }>;
  rejectRequest(request: FriendActionRequest): Promise<{ message: string }>;
  getFriends(offset?: number, limit?: number): Promise<{ friends: FriendInfo[] }>;
  getPendingRequests(offset?: number, limit?: number): Promise<{ requests: PendingRequestInfo[] }>;
  removeFriend(request: FriendActionRequest): Promise<{ message: string }>;
  getPendingRequestCount(): Promise<{ count: number }>;
}

export interface IAdminRepository {
  grantPoints(target_user_id: string, amount: number, description?: string): Promise<any>;
  deductPoints(target_user_id: string, amount: number, description?: string): Promise<any>;
  getAllUsers(offset?: number, limit?: number): Promise<{ users: User[]; total: number }>;
  getAllTransactions(offset?: number, limit?: number): Promise<{ transactions: Transaction[]; total: number }>;
  changeUserRole(target_user_id: string, new_role: string): Promise<any>;
  deactivateUser(target_user_id: string): Promise<any>;
}
