import { axiosInstance } from "@/infrastructure/api/client";

export interface DailyBonus {
  id: string;
  user_id: string;
  bonus_date: string;
  bonus_points: number;
  akerun_user_name: string;
  accessed_at?: string;
  created_at: string;
}

export interface TodayBonusResponse {
  claimed: boolean;
  bonus_points: number;
  total_days: number;
  daily_bonus?: DailyBonus;
}

export interface RecentBonusesResponse {
  bonuses: DailyBonus[];
  total_days: number;
}

export interface BonusSettingsResponse {
  bonus_points: number;
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

  // ボーナス設定を取得（管理者用）
  getBonusSettings: async (): Promise<BonusSettingsResponse> => {
    const response = await axiosInstance.get("/api/admin/bonus-settings");
    return response.data;
  },

  // ボーナス設定を更新（管理者用）
  updateBonusSettings: async (bonusPoints: number): Promise<void> => {
    await axiosInstance.put("/api/admin/bonus-settings", { bonus_points: bonusPoints });
  },
};
