import { axiosInstance } from "@/infrastructure/api/client";

export interface DailyBonus {
  id: string;
  user_id: string;
  bonus_date: string;
  login_completed: boolean;
  login_completed_at?: string;
  transfer_completed: boolean;
  transfer_completed_at?: string;
  exchange_completed: boolean;
  exchange_completed_at?: string;
  all_completed: boolean;
  all_completed_at?: string;
  total_bonus_points: number;
  completed_count: number;
  remaining_bonus: number;
}

export interface TodayBonusResponse {
  daily_bonus: DailyBonus;
  all_completed_count: number;
  can_claim_login_bonus: boolean;
  can_claim_transfer_bonus: boolean;
  can_claim_exchange_bonus: boolean;
}

export interface ClaimBonusResponse {
  bonus_awarded: number;
  new_balance: {
    id: string;
    username: string;
    balance: number;
  };
  daily_bonus: DailyBonus;
  message: string;
}

export interface RecentBonusesResponse {
  bonuses: DailyBonus[];
  all_completed_count: number;
}

export const dailyBonusApi = {
  // 本日のボーナス状況を取得
  getTodayBonus: async (): Promise<TodayBonusResponse> => {
    const response = await axiosInstance.get("/api/daily-bonus/today");
    return response.data;
  },

  // 最近のボーナス履歴を取得
  getRecentBonuses: async (limit: number = 7): Promise<RecentBonusesResponse> => {
    const response = await axiosInstance.get(`/api/daily-bonus/recent?limit=${limit}`);
    return response.data;
  },

  // ログインボーナスを請求
  claimLoginBonus: async (): Promise<ClaimBonusResponse> => {
    const response = await axiosInstance.post("/api/daily-bonus/claim-login");
    return response.data;
  },
};
