import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { getExchangeHistory, cancelExchange } from '../api/productsApi';
import type { ProductExchange } from '../types';

const STATUS_LABELS: Record<string, { label: string; color: string }> = {
  pending: { label: '処理中', color: 'bg-yellow-100 text-yellow-800' },
  completed: { label: '完了', color: 'bg-green-100 text-green-800' },
  cancelled: { label: 'キャンセル', color: 'bg-red-100 text-red-800' },
  delivered: { label: '配達済み', color: 'bg-blue-100 text-blue-800' },
};

export const ExchangeHistoryPage: React.FC = () => {
  const navigate = useNavigate();
  const [exchanges, setExchanges] = useState<ProductExchange[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadHistory();
  }, []);

  const loadHistory = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getExchangeHistory();
      setExchanges(data.Exchanges || []);
    } catch (err: any) {
      setError(err.response?.data?.error || '履歴の読み込みに失敗しました');
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = async (exchangeId: string) => {
    if (!confirm('この交換をキャンセルしますか？')) return;

    try {
      await cancelExchange(exchangeId);
      alert('キャンセルしました');
      loadHistory();
    } catch (err: any) {
      alert(err.response?.data?.error || 'キャンセルに失敗しました');
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString('ja-JP', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <div className="text-lg">読み込み中...</div>
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-4">交換履歴</h1>
        <p className="text-gray-600">これまでの商品交換履歴を確認できます</p>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 text-red-600 rounded-md">
          {error}
        </div>
      )}

      {exchanges.length === 0 ? (
        <div className="text-center py-12">
          <div className="text-gray-400 text-6xl mb-4">📦</div>
          <p className="text-gray-600 mb-4">まだ交換履歴がありません</p>
          <button
            onClick={() => navigate('/products')}
            className="text-blue-600 hover:underline"
          >
            商品一覧を見る
          </button>
        </div>
      ) : (
        <div className="space-y-4">
          {exchanges.map((exchange) => {
            const statusInfo = STATUS_LABELS[exchange.Status] || {
              label: exchange.Status,
              color: 'bg-gray-100 text-gray-800',
            };

            return (
              <div
                key={exchange.ID}
                className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow"
              >
                <div className="flex justify-between items-start mb-4">
                  <div>
                    <span
                      className={`inline-block px-3 py-1 text-sm font-semibold rounded ${statusInfo.color}`}
                    >
                      {statusInfo.label}
                    </span>
                  </div>
                  <div className="text-sm text-gray-500">
                    {formatDate(exchange.CreatedAt)}
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                  <div>
                    <div className="text-sm text-gray-600 mb-1">商品ID</div>
                    <div className="font-mono text-sm">{exchange.ProductID}</div>
                  </div>
                  <div>
                    <div className="text-sm text-gray-600 mb-1">数量</div>
                    <div className="font-semibold">{exchange.Quantity}個</div>
                  </div>
                  <div>
                    <div className="text-sm text-gray-600 mb-1">使用ポイント</div>
                    <div className="font-semibold text-blue-600">
                      {exchange.PointsUsed} pt
                    </div>
                  </div>
                  <div>
                    <div className="text-sm text-gray-600 mb-1">交換ID</div>
                    <div className="font-mono text-sm">{exchange.ID}</div>
                  </div>
                </div>

                {exchange.Notes && (
                  <div className="mb-4">
                    <div className="text-sm text-gray-600 mb-1">備考</div>
                    <div className="text-sm bg-gray-50 p-3 rounded">
                      {exchange.Notes}
                    </div>
                  </div>
                )}

                {exchange.DeliveredAt && (
                  <div className="text-sm text-gray-600">
                    配達完了: {formatDate(exchange.DeliveredAt)}
                  </div>
                )}

                {exchange.Status === 'completed' && !exchange.DeliveredAt && (
                  <div className="mt-4 pt-4 border-t border-gray-200">
                    <button
                      onClick={() => handleCancel(exchange.ID)}
                      className="text-red-600 hover:text-red-700 text-sm font-medium"
                    >
                      キャンセル
                    </button>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
