import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { PointRepository } from '@/infrastructure/api/repositories/PointRepository';
import { Transaction } from '@/core/domain/Transaction';
import { useAuthStore } from '@/shared/stores/authStore';

const pointRepository = new PointRepository();

export const HistoryPage: React.FC = () => {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const { user } = useAuthStore();
  const navigate = useNavigate();

  useEffect(() => {
    loadHistory();
  }, []);

  const loadHistory = async () => {
    try {
      const data = await pointRepository.getHistory(0, 50);
      setTransactions(data.transactions);
    } catch (error) {
      console.error('Failed to load history:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('ja-JP', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getTransactionInfo = (tx: Transaction) => {
    const isSent = tx.from_user_id === user?.id;
    const isReceived = tx.to_user_id === user?.id;

    // デイリーボーナス（from_user_idがnull）
    if (!tx.from_user_id && isReceived && tx.transaction_type === 'daily_bonus') {
      return {
        label: 'ボーナス',
        description: tx.description || 'デイリーボーナス',
        fromUser: 'システム',
        toUser: user?.display_name || user?.username || 'あなた',
        color: 'text-purple-600',
        sign: '+',
      };
    }

    // 送信の場合
    if (isSent) {
      const toUserName = tx.to_user?.display_name || tx.to_user?.username || '不明なユーザー';
      return {
        label: '送信',
        description: `${toUserName} へ送信`,
        fromUser: user?.display_name || user?.username || 'あなた',
        toUser: toUserName,
        color: 'text-red-600',
        sign: '-',
      };
    }

    // 受信の場合
    if (isReceived) {
      const fromUserName = tx.from_user?.display_name || tx.from_user?.username || 'システム';
      return {
        label: '受信',
        description: `${fromUserName} から受信`,
        fromUser: fromUserName,
        toUser: user?.display_name || user?.username || 'あなた',
        color: 'text-green-600',
        sign: '+',
      };
    }

    // その他（通常は発生しない）
    return {
      label: '取引',
      description: '取引',
      fromUser: tx.from_user?.display_name || 'システム',
      toUser: tx.to_user?.display_name || '不明',
      color: 'text-gray-600',
      sign: '',
    };
  };

  return (
    <div className="max-w-2xl mx-auto space-y-6 pb-20 md:pb-6">
      <div className="flex items-center mb-6">
        <button
          onClick={() => navigate(-1)}
          className="mr-4 p-2 hover:bg-gray-100 rounded-full"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <h1 className="text-2xl font-bold">取引履歴</h1>
      </div>

      {loading ? (
        <div className="flex justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
        </div>
      ) : transactions.length === 0 ? (
        <div className="bg-white rounded-xl shadow p-12 text-center">
          <svg className="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p className="text-gray-500">取引履歴がありません</p>
        </div>
      ) : (
        <div className="bg-white rounded-xl shadow divide-y divide-gray-200">
          {transactions.map((tx) => {
            const info = getTransactionInfo(tx);
            return (
              <div key={tx.id} className="p-4 hover:bg-gray-50 transition-colors">
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    {/* ラベルとステータス */}
                    <div className="flex items-center space-x-2 mb-2">
                      <span className={`text-xs px-2 py-1 rounded font-medium ${
                        info.label === '送信'
                          ? 'bg-red-50 text-red-700'
                          : info.label === 'ボーナス'
                          ? 'bg-purple-50 text-purple-700'
                          : 'bg-green-50 text-green-700'
                      }`}>
                        {info.label}
                      </span>
                      {tx.status === 'completed' ? (
                        <span className="text-xs px-2 py-1 bg-green-100 text-green-700 rounded">
                          完了
                        </span>
                      ) : (
                        <span className="text-xs px-2 py-1 bg-yellow-100 text-yellow-700 rounded">
                          {tx.status}
                        </span>
                      )}
                    </div>

                    {/* 送信者 → 受信者 */}
                    <div className="mb-2">
                      <div className="flex items-center text-sm text-gray-700">
                        <span className="font-medium">{info.fromUser}</span>
                        <svg className="w-4 h-4 mx-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
                        </svg>
                        <span className="font-medium">{info.toUser}</span>
                      </div>
                    </div>

                    {/* 説明文 */}
                    {tx.description && (
                      <div className="text-sm text-gray-500 mb-1">{tx.description}</div>
                    )}

                    {/* 日時 */}
                    <div className="text-xs text-gray-400">
                      {formatDate(tx.created_at)}
                    </div>
                  </div>

                  {/* 金額 */}
                  <div className="ml-4 text-right">
                    <div className={`text-xl font-bold ${info.color}`}>
                      {info.sign}{tx.amount.toLocaleString()} P
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
