import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowLeft, Settings, DoorOpen, Save } from "lucide-react";
import { dailyBonusApi } from "@/features/daily-bonus/api/dailyBonusApi";

export const AdminBonusSettingsPage: React.FC = () => {
    const navigate = useNavigate();
    const [bonusPoints, setBonusPoints] = useState<number>(5);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState<string | null>(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchSettings = async () => {
            try {
                setLoading(true);
                const data = await dailyBonusApi.getBonusSettings();
                setBonusPoints(data.bonus_points);
            } catch (err: any) {
                setError(err.response?.data?.error || "設定の取得に失敗しました");
            } finally {
                setLoading(false);
            }
        };
        fetchSettings();
    }, []);

    const handleSave = async () => {
        if (bonusPoints < 1) {
            setError("ポイント数は1以上を指定してください");
            return;
        }
        try {
            setSaving(true);
            setError(null);
            setMessage(null);
            await dailyBonusApi.updateBonusSettings(bonusPoints);
            setMessage("ボーナスポイント設定を更新しました");
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
                                    Akerun入退室ボーナス設定
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

                            <div className="space-y-6">
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 mb-2">
                                        <div className="flex items-center gap-2">
                                            <DoorOpen className="w-4 h-4" />
                                            入退室ボーナスポイント数
                                        </div>
                                    </label>
                                    <div className="flex items-center gap-3">
                                        <input
                                            type="number"
                                            min={1}
                                            max={1000}
                                            value={bonusPoints}
                                            onChange={(e) => setBonusPoints(parseInt(e.target.value) || 1)}
                                            className="w-32 px-3 py-2 border border-gray-300 rounded-lg text-lg font-bold text-center focus:outline-none focus:ring-2 focus:ring-purple-500"
                                        />
                                        <span className="text-lg font-medium text-gray-600">ポイント / 日</span>
                                    </div>
                                    <p className="mt-2 text-sm text-gray-500">
                                        Akerunで入退室したユーザーに1日1回付与されます（AM6:00リセット）
                                    </p>
                                </div>

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
                            <h3 className="font-bold text-blue-800 mb-2">ℹ️ Akerun入退室ボーナスについて</h3>
                            <ul className="text-sm text-blue-700 space-y-1">
                                <li>• サーバーが5分間隔でAkerun APIをポーリングします</li>
                                <li>• Akerunのユーザー名とアプリの氏名を照合してボーナスを自動付与</li>
                                <li>• 1ユーザーにつき1日1回まで（AM6:00 JST リセット）</li>
                                <li>• 環境変数 AKERUN_ACCESS_TOKEN, AKERUN_ORGANIZATION_ID の設定が必要です</li>
                            </ul>
                        </div>
                    </>
                )}
            </div>
        </div>
    );
};
