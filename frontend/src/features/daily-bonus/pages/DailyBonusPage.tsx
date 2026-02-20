import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowLeft, Calendar, TrendingUp } from "lucide-react";
import { dailyBonusApi, TodayBonusResponse, RecentBonusesResponse, DrawLotteryResponse } from "../api/dailyBonusApi";
import { DailyBonusCard } from "../components/DailyBonusCard";
import { LotteryAnimation } from "../components/LotteryAnimation";

export const DailyBonusPage = () => {
  const navigate = useNavigate();
  const [todayBonus, setTodayBonus] = useState<TodayBonusResponse | null>(null);
  const [recentBonuses, setRecentBonuses] = useState<RecentBonusesResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showLottery, setShowLottery] = useState(false);
  const [lotteryResult, setLotteryResult] = useState<DrawLotteryResponse | null>(null);

  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [todayData, recentData] = await Promise.all([
        dailyBonusApi.getTodayBonus(),
        dailyBonusApi.getRecentBonuses(14),
      ]);
      setTodayBonus(todayData);
      setRecentBonuses(recentData);

      // 未抽選のボーナスがある場合、ルーレットAPIを呼び出してアニメーション起動
      if (todayData.is_lottery_pending && todayData.daily_bonus) {
        try {
          const drawResult = await dailyBonusApi.drawLottery();
          setLotteryResult(drawResult);
          setShowLottery(true);
        } catch {
          // 抽選失敗時はデータ再取得
          const refreshedData = await dailyBonusApi.getTodayBonus();
          setTodayBonus(refreshedData);
        }
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "データの取得に失敗しました");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleLotteryComplete = async () => {
    setShowLottery(false);
    setLotteryResult(null);
    if (todayBonus?.daily_bonus) {
      try {
        await dailyBonusApi.markBonusViewed(todayBonus.daily_bonus.id);
        // データを再取得して最新状態を反映
        await fetchData();
      } catch {
        // mark-viewed失敗は無視
      }
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 p-4 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-purple-600 mx-auto mb-4"></div>
          <p className="text-gray-600">読み込み中...</p>
        </div>
      </div>
    );
  }

  if (error || !todayBonus || !recentBonuses) {
    return (
      <div className="min-h-screen bg-gray-50 p-4">
        <div className="max-w-md mx-auto">
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-center">
            <p className="text-red-800">{error || "データの読み込みに失敗しました"}</p>
            <button
              onClick={fetchData}
              className="mt-3 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
            >
              再試行
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      {/* 抽選アニメーション */}
      {showLottery && lotteryResult && (
        <LotteryAnimation
          result={lotteryResult}
          onComplete={handleLotteryComplete}
        />
      )}

      <div className="bg-white shadow-sm sticky top-0 z-10">
        <div className="max-w-4xl mx-auto px-4 py-4">
          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate("/")}
              className="p-2 hover:bg-gray-100 rounded-full transition-colors"
            >
              <ArrowLeft className="w-6 h-6" />
            </button>
            <h1 className="text-2xl font-bold">入退室ボーナス</h1>
          </div>
        </div>
      </div>

      <div className="max-w-4xl mx-auto px-4 py-6 space-y-6">
        {/* 今日のボーナスカード */}
        <DailyBonusCard todayBonus={todayBonus} />

        {/* 最近のボーナス履歴 */}
        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center gap-2 mb-4">
            <Calendar className="w-5 h-5 text-gray-600" />
            <h2 className="text-lg font-bold text-gray-800">最近のボーナス履歴</h2>
          </div>

          <div className="space-y-2">
            {recentBonuses.bonuses.length === 0 ? (
              <p className="text-center text-gray-500 py-8">履歴がありません</p>
            ) : (
              recentBonuses.bonuses.map((bonus) => (
                <div
                  key={bonus.id}
                  className="flex items-center justify-between p-3 rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors"
                >
                  <div>
                    <div className="font-medium text-gray-800">
                      {new Date(bonus.bonus_date).toLocaleDateString("ja-JP", {
                        month: "long",
                        day: "numeric",
                        weekday: "short",
                      })}
                    </div>
                    <div className="text-sm text-gray-500">
                      {bonus.lottery_tier_name && (
                        <span className="mr-2 px-1.5 py-0.5 bg-purple-100 text-purple-700 rounded text-xs">
                          {bonus.lottery_tier_name}
                        </span>
                      )}
                      {bonus.accessed_at
                        ? new Date(bonus.accessed_at).toLocaleTimeString("ja-JP", {
                          hour: "2-digit",
                          minute: "2-digit",
                        }) + " に入退室"
                        : ""}
                    </div>
                  </div>
                  <div className="font-bold text-purple-600">
                    +{bonus.bonus_points}P
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        {/* 統計 */}
        <div className="bg-gradient-to-br from-purple-600 to-pink-600 rounded-lg shadow-md p-6 text-white">
          <div className="flex items-center gap-2 mb-3">
            <TrendingUp className="w-5 h-5" />
            <h2 className="text-lg font-bold">統計</h2>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-white bg-opacity-20 rounded-lg p-3">
              <div className="text-sm opacity-90">累計獲得日数</div>
              <div className="text-3xl font-bold">{recentBonuses.total_days}</div>
            </div>
            <div className="bg-white bg-opacity-20 rounded-lg p-3">
              <div className="text-sm opacity-90">直近合計</div>
              <div className="text-3xl font-bold">
                {recentBonuses.bonuses.reduce((sum, b) => sum + b.bonus_points, 0)}P
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
