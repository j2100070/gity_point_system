import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowLeft, Settings, Save, Plus, Trash2 } from "lucide-react";
import { dailyBonusApi, LotteryTier, LotteryTierInput } from "@/features/daily-bonus/api/dailyBonusApi";

interface TierFormRow {
    name: string;
    points: number;
    probability: number;
    display_order: number;
}

export const AdminBonusSettingsPage: React.FC = () => {
    const navigate = useNavigate();
    const [tiers, setTiers] = useState<TierFormRow[]>([]);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState<string | null>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchSettings = async () => {
            try {
                setLoading(true);
                const data = await dailyBonusApi.getBonusSettings();
                if (data.lottery_tiers && data.lottery_tiers.length > 0) {
                    setTiers(
                        data.lottery_tiers.map((t: LotteryTier) => ({
                            name: t.name,
                            points: t.points,
                            probability: t.probability,
                            display_order: t.display_order,
                        }))
                    );
                } else {
                    // デフォルトティア
                    setTiers([
                        { name: "大当たり", points: 50, probability: 5, display_order: 1 },
                        { name: "当たり", points: 10, probability: 25, display_order: 2 },
                        { name: "小当たり", points: 5, probability: 50, display_order: 3 },
                        { name: "ハズレ", points: 0, probability: 20, display_order: 4 },
                    ]);
                }
            } catch (err: any) {
                setError(err.response?.data?.error || "設定の取得に失敗しました");
            } finally {
                setLoading(false);
            }
        };
        fetchSettings();
    }, []);

    const totalProbability = tiers.reduce((sum, t) => sum + t.probability, 0);

    const addTier = () => {
        setTiers([
            ...tiers,
            {
                name: "",
                points: 0,
                probability: 0,
                display_order: tiers.length + 1,
            },
        ]);
    };

    const removeTier = (index: number) => {
        setTiers(tiers.filter((_, i) => i !== index));
    };

    const updateTier = (index: number, field: keyof TierFormRow, value: string | number) => {
        setTiers(
            tiers.map((t, i) =>
                i === index ? { ...t, [field]: value } : t
            )
        );
    };

    const handleSave = async () => {
        // バリデーション
        for (const tier of tiers) {
            if (!tier.name.trim()) {
                setError("ティア名を入力してください");
                return;
            }
        }
        if (totalProbability <= 0) {
            setError("確率の合計は0%以上の値を指定してください");
            return;
        }

        try {
            setSaving(true);
            setError(null);
            setMessage(null);
            const tierInputs: LotteryTierInput[] = tiers.map((t, i) => ({
                name: t.name,
                points: t.points,
                probability: t.probability,
                display_order: i + 1,
            }));
            await dailyBonusApi.updateLotteryTiers(tierInputs);
            setMessage("抽選ティア設定を更新しました");
        } catch (err: any) {
            setError(err.response?.data?.error || "設定の更新に失敗しました");
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="min-h-screen bg-gray-50 pb-20">
            <div className="bg-white shadow-sm sticky top-0 z-10">
                <div className="max-w-4xl mx-auto px-4 py-4">
                    <div className="flex items-center gap-4">
                        <button
                            onClick={() => navigate("/admin")}
                            className="p-2 hover:bg-gray-100 rounded-full transition-colors"
                        >
                            <ArrowLeft className="w-6 h-6" />
                        </button>
                        <h1 className="text-2xl font-bold">ボーナス設定</h1>
                    </div>
                </div>
            </div>

            <div className="max-w-4xl mx-auto px-4 py-6 space-y-6">
                {loading ? (
                    <div className="text-center py-8">
                        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-purple-600 mx-auto mb-4"></div>
                        <p className="text-gray-600">読み込み中...</p>
                    </div>
                ) : (
                    <>
                        {/* 設定カード */}
                        <div className="bg-white rounded-lg shadow-md p-6">
                            <div className="flex items-center gap-2 mb-6">
                                <Settings className="w-5 h-5 text-gray-600" />
                                <h2 className="text-lg font-bold text-gray-800">
                                    🎰 抽選ティア設定
                                </h2>
                            </div>

                            {message && (
                                <div className="mb-4 p-3 bg-green-50 text-green-800 border border-green-200 rounded">
                                    {message}
                                </div>
                            )}

                            {error && (
                                <div className="mb-4 p-3 bg-red-50 text-red-800 border border-red-200 rounded">
                                    {error}
                                </div>
                            )}

                            <div className="space-y-4">
                                {/* ティアテーブル */}
                                <div className="overflow-x-auto">
                                    <table className="w-full">
                                        <thead>
                                            <tr className="border-b border-gray-200">
                                                <th className="text-left py-2 px-2 text-sm text-gray-600 font-medium">ティア名</th>
                                                <th className="text-left py-2 px-2 text-sm text-gray-600 font-medium">ポイント</th>
                                                <th className="text-left py-2 px-2 text-sm text-gray-600 font-medium">確率 (%)</th>
                                                <th className="py-2 px-2 w-12"></th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {tiers.map((tier, index) => (
                                                <tr key={index} className="border-b border-gray-100">
                                                    <td className="py-2 px-2">
                                                        <input
                                                            type="text"
                                                            value={tier.name}
                                                            onChange={(e) => updateTier(index, "name", e.target.value)}
                                                            className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
                                                            placeholder="ティア名"
                                                        />
                                                    </td>
                                                    <td className="py-2 px-2">
                                                        <input
                                                            type="number"
                                                            min={0}
                                                            value={tier.points}
                                                            onChange={(e) => updateTier(index, "points", parseInt(e.target.value) || 0)}
                                                            className="w-24 px-2 py-1.5 border border-gray-300 rounded text-sm text-right focus:outline-none focus:ring-2 focus:ring-purple-500"
                                                        />
                                                    </td>
                                                    <td className="py-2 px-2">
                                                        <input
                                                            type="number"
                                                            min={0}
                                                            max={100}
                                                            step={0.1}
                                                            value={tier.probability}
                                                            onChange={(e) => updateTier(index, "probability", parseFloat(e.target.value) || 0)}
                                                            className="w-24 px-2 py-1.5 border border-gray-300 rounded text-sm text-right focus:outline-none focus:ring-2 focus:ring-purple-500"
                                                        />
                                                    </td>
                                                    <td className="py-2 px-2 text-center">
                                                        <button
                                                            onClick={() => removeTier(index)}
                                                            className="p-1 text-red-400 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
                                                            title="削除"
                                                        >
                                                            <Trash2 className="w-4 h-4" />
                                                        </button>
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>

                                {/* 確率合計 */}
                                <div className={`text-sm font-medium px-2 ${totalProbability === 100
                                    ? "text-green-600"
                                    : totalProbability > 100
                                        ? "text-red-600"
                                        : "text-yellow-600"
                                    }`}>
                                    確率合計: {totalProbability.toFixed(1)}%
                                    {totalProbability < 100 && ` （残り${(100 - totalProbability).toFixed(1)}%はハズレになります）`}
                                    {totalProbability > 100 && " ⚠️ 100%を超えています"}
                                </div>

                                {/* ティア追加ボタン */}
                                <button
                                    onClick={addTier}
                                    className="flex items-center gap-2 px-4 py-2 text-purple-600 border border-purple-300 rounded-lg hover:bg-purple-50 transition-colors text-sm"
                                >
                                    <Plus className="w-4 h-4" />
                                    ティアを追加
                                </button>

                                {/* 保存ボタン */}
                                <button
                                    onClick={handleSave}
                                    disabled={saving}
                                    className="flex items-center gap-2 px-6 py-3 bg-purple-600 text-white rounded-lg hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
                                >
                                    <Save className="w-5 h-5" />
                                    {saving ? "保存中..." : "設定を保存"}
                                </button>
                            </div>
                        </div>

                        {/* 説明カード */}
                        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                            <h3 className="font-bold text-blue-800 mb-2">ℹ️ 抽選ボーナスについて</h3>
                            <ul className="text-sm text-blue-700 space-y-1">
                                <li>• サーバーが5分間隔でAkerun APIをポーリングします</li>
                                <li>• Akerunのユーザー名とアプリの氏名を照合してボーナスを自動付与</li>
                                <li>• 入退室時に上記ティア設定に基づいて抽選が行われます</li>
                                <li>• ユーザーはアプリでくじ引きアニメーションで結果を確認します</li>
                                <li>• 確率合計が100%未満の場合、残りはハズレ（0ポイント）です</li>
                                <li>• 1ユーザーにつき1日1回まで（AM6:00 JST リセット）</li>
                            </ul>
                        </div>
                    </>
                )}
            </div>
        </div>
    );
};
