import { CheckCircle, Circle, DoorOpen } from "lucide-react";
import { TodayBonusResponse } from "../api/dailyBonusApi";

interface DailyBonusCardProps {
  todayBonus: TodayBonusResponse;
}

export const DailyBonusCard: React.FC<DailyBonusCardProps> = ({ todayBonus }) => {
  const { claimed, total_days, daily_bonus } = todayBonus;

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <DoorOpen className="w-6 h-6 text-purple-600" />
          <h2 className="text-xl font-bold text-gray-800">入退室ボーナス</h2>
        </div>
        <div className="text-right">
          <div className="text-sm text-gray-600">🎰 抽選制</div>
        </div>
      </div>

      {/* 今日の状態 */}
      <div
        className={`flex items-center justify-between p-4 rounded-lg mb-4 ${claimed
          ? "bg-green-50 border border-green-200"
          : "bg-gray-50 border border-gray-200"
          }`}
      >
        <div className="flex items-center gap-3">
          {claimed ? (
            <CheckCircle className="w-6 h-6 text-green-600" />
          ) : (
            <Circle className="w-6 h-6 text-gray-400" />
          )}
          <div>
            <span
              className={`font-medium text-lg ${claimed ? "text-green-800" : "text-gray-600"
                }`}
            >
              {claimed ? "本日のボーナス獲得済み" : "Akerunで入退室するとボーナス獲得"}
            </span>
            {claimed && daily_bonus && (
              <div className="text-sm text-green-600 mt-1">
                +{daily_bonus.bonus_points}P 獲得
                {daily_bonus.lottery_tier_name && (
                  <span className="ml-2 px-2 py-0.5 bg-purple-100 text-purple-700 rounded-full text-xs font-medium">
                    {daily_bonus.lottery_tier_name}
                  </span>
                )}
                {daily_bonus.accessed_at && (
                  <span className="ml-2 text-gray-500">
                    ({new Date(daily_bonus.accessed_at).toLocaleTimeString("ja-JP", {
                      hour: "2-digit",
                      minute: "2-digit",
                    })})
                  </span>
                )}
              </div>
            )}
          </div>
        </div>
      </div>

      {!claimed && (
        <div className="bg-purple-50 border border-purple-200 p-3 rounded-lg text-center">
          <div className="text-sm text-purple-700">
            Akerunで入退室すると <span className="font-bold">🎰 抽選</span> でポイント獲得！
          </div>
          <div className="text-xs text-purple-500 mt-1">
            ※ 1日1回まで（AM6:00リセット）
          </div>
        </div>
      )}

      {claimed && (
        <div className="bg-gradient-to-r from-purple-600 to-pink-600 text-white p-4 rounded-lg text-center">
          <div className="text-lg font-bold">🎉 本日のボーナス獲得済み！</div>
          <div className="text-sm opacity-90 mt-1">
            明日もAkerunで入退室してボーナスをゲットしよう
          </div>
        </div>
      )}

      <div className="mt-4 pt-4 border-t border-gray-200 text-center">
        <div className="text-sm text-gray-600">
          累計獲得日数:{" "}
          <span className="font-bold text-purple-600">{total_days}日</span>
        </div>
      </div>
    </div>
  );
};
