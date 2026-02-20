import { useState, useEffect, useCallback } from "react";
import { DrawLotteryResponse } from "../api/dailyBonusApi";

interface LotteryAnimationProps {
    result: DrawLotteryResponse;
    onComplete: () => void;
}

const EMOJIS = ["🎰", "💎", "⭐", "🎯", "🔥", "✨", "🎊", "🎁"];

export const LotteryAnimation: React.FC<LotteryAnimationProps> = ({
    result,
    onComplete,
}) => {
    const [phase, setPhase] = useState<"spinning" | "reveal" | "done">("spinning");
    const [displayPoints, setDisplayPoints] = useState(0);
    const [currentEmoji, setCurrentEmoji] = useState("🎰");

    // スロット風の数値スピン
    useEffect(() => {
        if (phase !== "spinning") return;
        const interval = setInterval(() => {
            setDisplayPoints(Math.floor(Math.random() * 50) + 1);
            setCurrentEmoji(EMOJIS[Math.floor(Math.random() * EMOJIS.length)]);
        }, 80);

        const timeout = setTimeout(() => {
            clearInterval(interval);
            setPhase("reveal");
        }, 2000);

        return () => {
            clearInterval(interval);
            clearTimeout(timeout);
        };
    }, [phase]);

    // 結果の表示
    useEffect(() => {
        if (phase !== "reveal") return;
        setDisplayPoints(result.bonus_points);

        const timer = setTimeout(() => {
            setPhase("done");
        }, 1500);
        return () => clearTimeout(timer);
    }, [phase, result.bonus_points]);

    const handleClose = useCallback(() => {
        onComplete();
    }, [onComplete]);

    const isWin = result.bonus_points > 0;
    const tierName = result.lottery_tier_name || "通常";

    return (
        <div
            className="fixed inset-0 z-50 flex items-center justify-center"
            style={{ backgroundColor: "rgba(0, 0, 0, 0.75)" }}
        >
            <div
                className="relative max-w-sm w-full mx-4 rounded-2xl overflow-hidden shadow-2xl"
                style={{
                    background: isWin
                        ? "linear-gradient(135deg, #6b21a8, #db2777, #f59e0b)"
                        : "linear-gradient(135deg, #374151, #4b5563, #6b7280)",
                }}
            >
                {/* パーティクルエフェクト */}
                {phase === "reveal" && isWin && (
                    <div className="absolute inset-0 overflow-hidden pointer-events-none">
                        {Array.from({ length: 20 }).map((_, i) => (
                            <div
                                key={i}
                                className="absolute w-2 h-2 rounded-full"
                                style={{
                                    left: `${Math.random() * 100}%`,
                                    top: `${Math.random() * 100}%`,
                                    backgroundColor: ["#fbbf24", "#f472b6", "#a78bfa", "#34d399"][i % 4],
                                    animation: `particle ${1 + Math.random()}s ease-out forwards`,
                                    animationDelay: `${Math.random() * 0.5}s`,
                                    opacity: 0.8,
                                }}
                            />
                        ))}
                    </div>
                )}

                <div className="p-8 text-center text-white relative">
                    {/* タイトル */}
                    <h2
                        className="text-xl font-bold mb-6"
                        style={{ textShadow: "0 2px 8px rgba(0,0,0,0.3)" }}
                    >
                        🎰 入退室ボーナス抽選
                    </h2>

                    {/* スロット表示 */}
                    <div
                        className="bg-white bg-opacity-15 backdrop-blur-sm rounded-xl p-6 mb-6"
                        style={{ border: "2px solid rgba(255,255,255,0.2)" }}
                    >
                        {phase === "spinning" ? (
                            <>
                                <div className="text-5xl mb-3" style={{
                                    animation: "pulse 0.3s ease-in-out infinite",
                                }}>
                                    {currentEmoji}
                                </div>
                                <div
                                    className="text-4xl font-bold tabular-nums"
                                    style={{
                                        fontFamily: '"SF Mono", "Roboto Mono", monospace',
                                        textShadow: "0 0 20px rgba(255,255,255,0.5)",
                                    }}
                                >
                                    {displayPoints}P
                                </div>
                                <div className="text-sm opacity-70 mt-2">抽選中...</div>
                            </>
                        ) : (
                            <>
                                <div className="text-5xl mb-3" style={{
                                    animation: "bounce 0.6s ease-out",
                                }}>
                                    {isWin ? "🎊" : "😢"}
                                </div>
                                <div
                                    className="text-4xl font-bold"
                                    style={{
                                        fontFamily: '"SF Mono", "Roboto Mono", monospace',
                                        animation: "scaleIn 0.4s ease-out",
                                        textShadow: "0 0 20px rgba(255,255,255,0.5)",
                                    }}
                                >
                                    {isWin ? `+${result.bonus_points}P` : "0P"}
                                </div>
                                <div
                                    className="text-lg font-medium mt-2"
                                    style={{
                                        animation: "fadeIn 0.5s ease-out 0.3s both",
                                    }}
                                >
                                    {tierName}
                                </div>
                            </>
                        )}
                    </div>

                    {/* 閉じるボタン */}
                    {phase !== "spinning" && (
                        <button
                            onClick={handleClose}
                            className="px-8 py-3 bg-white text-purple-800 rounded-full font-bold text-lg shadow-lg hover:shadow-xl transition-all"
                            style={{
                                animation: "fadeIn 0.5s ease-out 0.5s both",
                            }}
                        >
                            OK
                        </button>
                    )}
                </div>
            </div>

            {/* CSS Animations */}
            <style>{`
        @keyframes particle {
          0% { transform: scale(0) translateY(0); opacity: 1; }
          100% { transform: scale(1.5) translateY(-80px); opacity: 0; }
        }
        @keyframes pulse {
          0%, 100% { transform: scale(1); }
          50% { transform: scale(1.15); }
        }
        @keyframes bounce {
          0% { transform: scale(0.3); opacity: 0; }
          50% { transform: scale(1.1); }
          70% { transform: scale(0.95); }
          100% { transform: scale(1); opacity: 1; }
        }
        @keyframes scaleIn {
          0% { transform: scale(0.5); opacity: 0; }
          100% { transform: scale(1); opacity: 1; }
        }
        @keyframes fadeIn {
          0% { opacity: 0; transform: translateY(10px); }
          100% { opacity: 1; transform: translateY(0); }
        }
      `}</style>
        </div>
    );
};
