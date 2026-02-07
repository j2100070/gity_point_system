import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuthStore } from '@/shared/stores/authStore';
import { PointRepository } from '@/infrastructure/api/repositories/PointRepository';

const pointRepository = new PointRepository();

export const DashboardPage: React.FC = () => {
  const { user } = useAuthStore();
  const [balance, setBalance] = useState<number>(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadBalance();
  }, []);

  const loadBalance = async () => {
    try {
      const data = await pointRepository.getBalance();
      setBalance(data.balance);
    } catch (error) {
      console.error('Failed to load balance:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6 pb-20 md:pb-6">
      {/* 残高カード */}
      <div className="bg-gradient-to-br from-primary-500 to-primary-600 rounded-2xl shadow-lg p-6 text-white">
        <div className="text-sm opacity-90 mb-2">現在の残高</div>
        {loading ? (
          <div className="h-12 flex items-center">
            <div className="animate-pulse text-2xl">読み込み中...</div>
          </div>
        ) : (
          <div className="text-4xl font-bold mb-4">{balance.toLocaleString()} P</div>
        )}
        <div className="text-sm opacity-75">
          {user?.display_name} さん
        </div>
      </div>

      {/* アクションボタン */}
      <div className="grid grid-cols-2 gap-4">
        <Link
          to="/qr/receive"
          className="bg-white rounded-xl shadow p-6 flex flex-col items-center justify-center hover:shadow-md transition-shadow"
        >
          <div className="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center mb-3">
            <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
          </div>
          <div className="text-sm font-medium text-gray-900">受け取る</div>
        </Link>

        <Link
          to="/qr/send"
          className="bg-white rounded-xl shadow p-6 flex flex-col items-center justify-center hover:shadow-md transition-shadow"
        >
          <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center mb-3">
            <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
            </svg>
          </div>
          <div className="text-sm font-medium text-gray-900">送る</div>
        </Link>

        <Link
          to="/qr/scan"
          className="bg-white rounded-xl shadow p-6 flex flex-col items-center justify-center hover:shadow-md transition-shadow"
        >
          <div className="w-12 h-12 bg-purple-100 rounded-full flex items-center justify-center mb-3">
            <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
            </svg>
          </div>
          <div className="text-sm font-medium text-gray-900">スキャン</div>
        </Link>

        <Link
          to="/transfer"
          className="bg-white rounded-xl shadow p-6 flex flex-col items-center justify-center hover:shadow-md transition-shadow"
        >
          <div className="w-12 h-12 bg-orange-100 rounded-full flex items-center justify-center mb-3">
            <svg className="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
            </svg>
          </div>
          <div className="text-sm font-medium text-gray-900">送金</div>
        </Link>
      </div>

      {/* クイックリンク */}
      <div className="bg-white rounded-xl shadow p-4">
        <h3 className="text-sm font-medium text-gray-900 mb-3">メニュー</h3>
        <div className="space-y-2">
          <Link
            to="/history"
            className="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <div className="flex items-center">
              <svg className="w-5 h-5 text-gray-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-sm text-gray-700">取引履歴</span>
            </div>
            <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </Link>

          <Link
            to="/friends"
            className="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <div className="flex items-center">
              <svg className="w-5 h-5 text-gray-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
              <span className="text-sm text-gray-700">友達管理</span>
            </div>
            <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </Link>

          <Link
            to="/products"
            className="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <div className="flex items-center">
              <svg className="w-5 h-5 text-gray-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
              </svg>
              <span className="text-sm text-gray-700">商品交換</span>
            </div>
            <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </Link>

          {user?.role === 'admin' && (
            <Link
              to="/admin"
              className="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <div className="flex items-center">
                <svg className="w-5 h-5 text-gray-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                <span className="text-sm text-gray-700">管理者画面</span>
              </div>
              <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
            </Link>
          )}
        </div>
      </div>
    </div>
  );
};
