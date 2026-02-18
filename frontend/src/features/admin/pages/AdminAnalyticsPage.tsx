import React, { useEffect, useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { AdminRepository } from '@/infrastructure/api/repositories/AdminRepository';
import {
    LineChart,
    Line,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    Legend,
    ResponsiveContainer,
} from 'recharts';

const adminRepository = new AdminRepository();

interface AnalyticsData {
    summary: {
        total_points_in_circulation: number;
        average_balance: number;
        points_issued_this_month: number;
        transactions_this_month: number;
        active_users: number;
    };
    top_holders: {
        user_id: string;
        username: string;
        display_name: string;
        balance: number;
        percentage: number;
    }[];
    daily_stats: {
        date: string;
        issued: number;
        consumed: number;
        transferred: number;
    }[];
    transaction_type_breakdown: {
        type: string;
        count: number;
        total_amount: number;
    }[];
}

const PERIOD_OPTIONS = [
    { value: 7, label: '7日間' },
    { value: 30, label: '30日間' },
    { value: 90, label: '90日間' },
];

const getTypeLabel = (type: string) => {
    switch (type) {
        case 'transfer': return 'ユーザー間送金';
        case 'admin_grant': return '管理者付与';
        case 'admin_deduct': return '管理者減算';
        case 'system_grant': return 'システム付与';
        case 'system_expire': return 'ポイント期限切れ';
        default: return type;
    }
};

export const AdminAnalyticsPage: React.FC = () => {
    const [data, setData] = useState<AnalyticsData | null>(null);
    const [loading, setLoading] = useState(true);
    const [days, setDays] = useState(30);
    const navigate = useNavigate();

    useEffect(() => {
        loadAnalytics();
    }, [days]);

    const loadAnalytics = async () => {
        setLoading(true);
        try {
            const result = await adminRepository.getAnalytics(days);
            setData(result);
        } catch (error) {
            console.error('Failed to load analytics:', error);
        } finally {
            setLoading(false);
        }
    };

    if (loading || !data) {
        return (
            <div className="max-w-7xl mx-auto pb-20 md:pb-6">
                <div className="flex justify-center py-20">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
                </div>
            </div>
        );
    }

    return (
        <div className="max-w-7xl mx-auto space-y-6 pb-20 md:pb-6">
            {/* ヘッダー */}
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center">
                    <button
                        onClick={() => navigate(-1)}
                        className="mr-4 p-2 hover:bg-gray-100 rounded-full"
                    >
                        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                        </svg>
                    </button>
                    <h1 className="text-2xl font-bold">分析ダッシュボード</h1>
                </div>
            </div>

            {/* KPIカード */}
            <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
                <div className="bg-white rounded-xl shadow p-4">
                    <div className="text-xs text-gray-500 mb-1">ポイント総流通量</div>
                    <div className="text-2xl font-bold text-gray-900">
                        {data.summary.total_points_in_circulation.toLocaleString()}
                    </div>
                    <div className="text-xs text-gray-400">P</div>
                </div>
                <div className="bg-white rounded-xl shadow p-4">
                    <div className="text-xs text-gray-500 mb-1">平均保有ポイント</div>
                    <div className="text-2xl font-bold text-gray-900">
                        {Math.round(data.summary.average_balance).toLocaleString()}
                    </div>
                    <div className="text-xs text-gray-400">P</div>
                </div>
                <div className="bg-white rounded-xl shadow p-4">
                    <div className="text-xs text-gray-500 mb-1">今月の発行量</div>
                    <div className="text-2xl font-bold text-green-600">
                        {data.summary.points_issued_this_month.toLocaleString()}
                    </div>
                    <div className="text-xs text-gray-400">P</div>
                </div>
                <div className="bg-white rounded-xl shadow p-4">
                    <div className="text-xs text-gray-500 mb-1">今月のTx数</div>
                    <div className="text-2xl font-bold text-blue-600">
                        {data.summary.transactions_this_month.toLocaleString()}
                    </div>
                    <div className="text-xs text-gray-400">件</div>
                </div>
                <div className="bg-white rounded-xl shadow p-4 col-span-2 md:col-span-1">
                    <div className="text-xs text-gray-500 mb-1">アクティブユーザー</div>
                    <div className="text-2xl font-bold text-purple-600">
                        {data.summary.active_users.toLocaleString()}
                    </div>
                    <div className="text-xs text-gray-400">人</div>
                </div>
            </div>

            {/* 日別推移チャート */}
            <div className="bg-white rounded-xl shadow p-6">
                <div className="flex items-center justify-between mb-4">
                    <h2 className="text-lg font-semibold text-gray-900">日別ポイント推移</h2>
                    <select
                        value={days}
                        onChange={(e) => setDays(Number(e.target.value))}
                        className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                    >
                        {PERIOD_OPTIONS.map((opt) => (
                            <option key={opt.value} value={opt.value}>
                                過去{opt.label}
                            </option>
                        ))}
                    </select>
                </div>
                {data.daily_stats.length > 0 ? (
                    <ResponsiveContainer width="100%" height={320}>
                        <LineChart data={data.daily_stats}>
                            <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                            <XAxis
                                dataKey="date"
                                tickFormatter={(v) => {
                                    const d = new Date(v);
                                    return `${d.getMonth() + 1}/${d.getDate()}`;
                                }}
                                fontSize={12}
                                stroke="#9ca3af"
                            />
                            <YAxis fontSize={12} stroke="#9ca3af" />
                            <Tooltip
                                labelFormatter={(v) => {
                                    const d = new Date(v as string);
                                    return `${d.getFullYear()}/${d.getMonth() + 1}/${d.getDate()}`;
                                }}
                                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                                formatter={((value: any, name: any) => {
                                    const labels: Record<string, string> = {
                                        issued: '発行',
                                        consumed: '消費',
                                        transferred: '送金',
                                    };
                                    return [`${Number(value).toLocaleString()} P`, labels[name] || name];
                                }) as any}
                            />
                            <Legend
                                formatter={(value) => {
                                    const labels: Record<string, string> = {
                                        issued: '発行',
                                        consumed: '消費',
                                        transferred: '送金',
                                    };
                                    return labels[value] || value;
                                }}
                            />
                            <Line
                                type="monotone"
                                dataKey="issued"
                                stroke="#22c55e"
                                strokeWidth={2}
                                dot={false}
                                activeDot={{ r: 4 }}
                            />
                            <Line
                                type="monotone"
                                dataKey="consumed"
                                stroke="#ef4444"
                                strokeWidth={2}
                                dot={false}
                                activeDot={{ r: 4 }}
                            />
                            <Line
                                type="monotone"
                                dataKey="transferred"
                                stroke="#3b82f6"
                                strokeWidth={2}
                                dot={false}
                                activeDot={{ r: 4 }}
                            />
                        </LineChart>
                    </ResponsiveContainer>
                ) : (
                    <div className="flex items-center justify-center h-64 text-gray-400">
                        <p>この期間のデータがありません</p>
                    </div>
                )}
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* ポイント保有ランキング */}
                <div className="bg-white rounded-xl shadow">
                    <div className="p-4 border-b border-gray-200 flex items-center justify-between">
                        <h2 className="text-lg font-semibold text-gray-900">ポイント保有ランキング</h2>
                        <Link to="/admin/users" className="text-sm text-primary-600 hover:underline">
                            全ユーザー →
                        </Link>
                    </div>
                    <div className="divide-y divide-gray-100">
                        {data.top_holders.map((holder, index) => (
                            <div key={holder.user_id} className="flex items-center px-4 py-3">
                                <div className="w-8 text-center font-bold text-gray-400">{index + 1}</div>
                                <div className="flex-1 ml-3">
                                    <div className="font-medium text-gray-900">{holder.display_name}</div>
                                    <div className="text-xs text-gray-500">@{holder.username}</div>
                                </div>
                                <div className="text-right">
                                    <div className="font-semibold text-gray-900">
                                        {holder.balance.toLocaleString()} P
                                    </div>
                                    <div className="text-xs text-gray-400">{holder.percentage.toFixed(1)}%</div>
                                </div>
                            </div>
                        ))}
                        {data.top_holders.length === 0 && (
                            <div className="text-center py-8 text-gray-400">データがありません</div>
                        )}
                    </div>
                </div>

                {/* トランザクション種別構成 */}
                <div className="bg-white rounded-xl shadow">
                    <div className="p-4 border-b border-gray-200 flex items-center justify-between">
                        <h2 className="text-lg font-semibold text-gray-900">トランザクション種別構成</h2>
                        <Link to="/admin/transactions" className="text-sm text-primary-600 hover:underline">
                            全トランザクション →
                        </Link>
                    </div>
                    <div className="p-4">
                        <table className="w-full">
                            <thead>
                                <tr className="text-xs text-gray-500 border-b">
                                    <th className="text-left py-2">種別</th>
                                    <th className="text-right py-2">件数</th>
                                    <th className="text-right py-2">金額合計</th>
                                </tr>
                            </thead>
                            <tbody>
                                {data.transaction_type_breakdown.map((item) => (
                                    <tr
                                        key={item.type}
                                        className="border-b border-gray-50 cursor-pointer hover:bg-gray-50 transition-colors"
                                        onClick={() => navigate(`/admin/transactions?type=${item.type}`)}
                                    >
                                        <td className="py-3 text-sm font-medium text-primary-600 hover:underline">
                                            {getTypeLabel(item.type)}
                                        </td>
                                        <td className="py-3 text-sm text-right text-gray-600">
                                            {item.count.toLocaleString()} 件
                                        </td>
                                        <td className="py-3 text-sm text-right font-medium text-gray-900">
                                            {item.total_amount.toLocaleString()} P
                                        </td>
                                    </tr>
                                ))}
                                {data.transaction_type_breakdown.length === 0 && (
                                    <tr>
                                        <td colSpan={3} className="text-center py-8 text-gray-400">
                                            データがありません
                                        </td>
                                    </tr>
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    );
};
