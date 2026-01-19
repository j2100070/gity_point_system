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

  const getTransactionLabel = (tx: Transaction) => {
    if (tx.from_user_id === user?.id) {
      return {
        label: '送信',
        name: tx.to_user?.display_name || '不明',
        color: 'text-red-600',
        sign: '-',
      };
    } else {
      return {
        label: '受信',
        name: tx.from_user?.display_name || 'システム',
        color: 'text-green-600',
        sign: '+',
      };
    }
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
            const info = getTransactionLabel(tx);
            return (
              <div key={tx.id} className="p-4 hover:bg-gray-50">
                <div className="flex justify-between items-start mb-2">
                  <div className="flex-1">
                    <div className="flex items-center space-x-2 mb-1">
                      <span className="text-xs px-2 py-1 bg-gray-100 text-gray-700 rounded">
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
                    <div className="font-medium text-gray-900">{info.name}</div>
                    {tx.description && (
                      <div className="text-sm text-gray-500 mt-1">{tx.description}</div>
                    )}
                    <div className="text-xs text-gray-400 mt-1">
                      {formatDate(tx.created_at)}
                    </div>
                  </div>
                  <div className={`text-lg font-bold ${info.color}`}>
                    {info.sign}{tx.amount.toLocaleString()} P
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
