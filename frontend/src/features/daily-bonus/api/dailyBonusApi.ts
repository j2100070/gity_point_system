import { axiosInstance } from "@/infrastructure/api/client";

export interface DailyBonus {
  id: string;
  user_id: string;
  bonus_date: string;
  bonus_points: number;
  akerun_user_name: string;
  accessed_at?: string;
  lottery_tier_name?: string;
  is_viewed: boolean;
  is_drawn: boolean;
  created_at: string;
}

export interface DrawLotteryResponse {
  bonus_points: number;
  lottery_tier_name: string;
  bonus_id: string;
}

export interface TodayBonusResponse {
  claimed: boolean;
  bonus_points: number;
  total_days: number;
  daily_bonus?: DailyBonus;
  is_lottery_pending: boolean;
}

export interface RecentBonusesResponse {
  bonuses: DailyBonus[];
  total_days: number;
}

export interface LotteryTier {
  id: string;
  name: string;
  points: number;
  probability: number;
  display_order: number;
  is_active: boolean;
}

export interface BonusSettingsResponse {
  bonus_points: number;
  lottery_tiers: LotteryTier[];
}

export interface LotteryTierInput {
  name: string;
  points: number;
  probability: number;
  display_order: number;
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

  // ボーナスを閲覧済みにする
  markBonusViewed: async (bonusId: string): Promise<void> => {
    await axiosInstance.post("/api/daily-bonus/mark-viewed", { bonus_id: bonusId });
  },

  // ルーレットを実行しポイントを付与する
  drawLottery: async (): Promise<DrawLotteryResponse> => {
    const response = await axiosInstance.post("/api/daily-bonus/draw");
    return response.data;
  },

  // ボーナス設定を取得（管理者用）
  getBonusSettings: async (): Promise<BonusSettingsResponse> => {
    const response = await axiosInstance.get("/api/admin/bonus-settings");
    return response.data;
  },

  // 抽選ティアを更新（管理者用）
  updateLotteryTiers: async (tiers: LotteryTierInput[]): Promise<void> => {
    await axiosInstance.put("/api/admin/lottery-tiers", { tiers });
  },
};
