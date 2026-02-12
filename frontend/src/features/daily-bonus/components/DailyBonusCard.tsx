import { useState } from "react";
import { CheckCircle, Circle, Gift } from "lucide-react";
import { dailyBonusApi, DailyBonus } from "../api/dailyBonusApi";
import { useAuthStore } from "@/shared/stores/authStore";

interface DailyBonusCardProps {
  dailyBonus: DailyBonus;
  allCompletedCount: number;
  canClaimLoginBonus: boolean;
  onBonusClaimed?: () => void;
}

export const DailyBonusCard: React.FC<DailyBonusCardProps> = ({
  dailyBonus,
  allCompletedCount,
  canClaimLoginBonus,
  onBonusClaimed,
}) => {
  const [claiming, setClaiming] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const { updateUserBalance } = useAuthStore();

  const handleClaimLogin = async () => {
    try {
      setClaiming(true);
      setMessage(null);
      const response = await dailyBonusApi.claimLoginBonus();
      setMessage(response.message);

      // Update user balance in store
      updateUserBalance(response.new_balance.balance);

      onBonusClaimed?.();
    } catch (error: any) {
      setMessage(error.response?.data?.error || "ボーナスの取得に失敗しました");
    } finally {
      setClaiming(false);
    }
  };

  const bonusItems = [
    {
      label: "ログイン",
      completed: dailyBonus.login_completed,
      points: 10,
      canClaim: canClaimLoginBonus,
      onClaim: handleClaimLogin,
    },
    {
      label: "送金",
      completed: dailyBonus.transfer_completed,
      points: 10,
      canClaim: false, // Auto-triggered
    },
    {
      label: "商品交換",
      completed: dailyBonus.exchange_completed,
      points: 10,
      canClaim: false, // Auto-triggered
    },
  ];

  const allComplete = dailyBonus.all_completed;
  const totalPossiblePoints = 50;

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Gift className="w-6 h-6 text-purple-600" />
          <h2 className="text-xl font-bold text-gray-800">デイリーボーナス</h2>
        </div>
        <div className="text-right">
          <div className="text-sm text-gray-600">獲得ポイント</div>
          <div className="text-2xl font-bold text-purple-600">
            {dailyBonus.total_bonus_points}P
          </div>
        </div>
      </div>

      {message && (
        <div
          className={`mb-4 p-3 rounded ${
            message.includes("失敗") || message.includes("既に")
              ? "bg-yellow-50 text-yellow-800 border border-yellow-200"
              : "bg-green-50 text-green-800 border border-green-200"
          }`}
        >
          {message}
        </div>
      )}

      <div className="space-y-3 mb-4">
        {bonusItems.map((item, index) => (
          <div
            key={index}
            className={`flex items-center justify-between p-3 rounded-lg ${
              item.completed
                ? "bg-green-50 border border-green-200"
                : "bg-gray-50 border border-gray-200"
            }`}
          >
            <div className="flex items-center gap-3">
              {item.completed ? (
                <CheckCircle className="w-5 h-5 text-green-600" />
              ) : (
                <Circle className="w-5 h-5 text-gray-400" />
              )}
              <span
                className={`font-medium ${
                  item.completed ? "text-green-800" : "text-gray-600"
                }`}
              >
                {item.label}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-sm font-semibold text-gray-700">
                +{item.points}P
              </span>
              {item.canClaim && !item.completed && (
                <button
                  onClick={item.onClaim}
                  disabled={claiming}
                  className="px-3 py-1 bg-purple-600 text-white text-sm rounded hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {claiming ? "取得中..." : "取得"}
                </button>
              )}
            </div>
          </div>
        ))}
      </div>

      {allComplete && (
        <div className="bg-gradient-to-r from-purple-600 to-pink-600 text-white p-4 rounded-lg text-center">
          <div className="text-lg font-bold mb-1">全達成ボーナス!</div>
          <div className="text-sm">本日のボーナスを全て獲得しました +50P</div>
        </div>
      )}

      {!allComplete && (
        <div className="bg-gray-100 p-3 rounded-lg text-center">
          <div className="text-sm text-gray-600">
            全達成で合計 <span className="font-bold text-purple-600">{totalPossiblePoints}P</span> 獲得!
          </div>
          <div className="text-xs text-gray-500 mt-1">
            残り {dailyBonus.remaining_bonus}P
          </div>
        </div>
      )}

      <div className="mt-4 pt-4 border-t border-gray-200 text-center">
        <div className="text-sm text-gray-600">
          累計全達成回数:{" "}
          <span className="font-bold text-purple-600">{allCompletedCount}回</span>
        </div>
      </div>
    </div>
  );
};
