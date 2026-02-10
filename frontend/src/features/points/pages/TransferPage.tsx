import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { PointRepository } from '@/infrastructure/api/repositories/PointRepository';
import { FriendshipRepository } from '@/infrastructure/api/repositories/FriendshipRepository';
import { FriendInfo } from '@/core/domain/Friendship';

const pointRepository = new PointRepository();
const friendshipRepository = new FriendshipRepository();

interface LocationState {
  friendId?: string;
  friendName?: string;
  friendUsername?: string;
}

export const TransferPage: React.FC = () => {
  const [toUserId, setToUserId] = useState('');
  const [selectedFriendName, setSelectedFriendName] = useState('');
  const [amount, setAmount] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState('');
  const [friends, setFriends] = useState<FriendInfo[]>([]);
  const [friendsLoading, setFriendsLoading] = useState(true);
  const navigate = useNavigate();
  const location = useLocation();

  // フレンドリストから遷移した場合、送信先を自動設定
  useEffect(() => {
    const state = location.state as LocationState | null;
    if (state?.friendId) {
      setToUserId(state.friendId);
      setSelectedFriendName(state.friendName || '');
    }
  }, [location.state]);

  // フレンドリストを読み込み
  useEffect(() => {
    loadFriends();
  }, []);

  const loadFriends = async () => {
    try {
      const data = await friendshipRepository.getFriends();
      setFriends(data.friends || []);
    } catch (error) {
      console.error('Failed to load friends:', error);
    } finally {
      setFriendsLoading(false);
    }
  };

  const generateIdempotencyKey = () => {
    return `transfer-${Date.now()}-${Math.random().toString(36).substring(7)}`;
  };

  const handleSelectFriend = (item: FriendInfo) => {
    if (item.friend) {
      setToUserId(item.friend.id);
      setSelectedFriendName(item.friend.display_name);
    }
  };

  const handleTransfer = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await pointRepository.transfer({
        to_user_id: toUserId.trim(),
        amount: parseInt(amount),
        idempotency_key: generateIdempotencyKey(),
        description: description.trim() || undefined,
      });
      setSuccess(true);
      setTimeout(() => {
        navigate('/dashboard');
      }, 2000);
    } catch (err: any) {
      setError(err.response?.data?.error || 'ポイント送信に失敗しました');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-md mx-auto space-y-6 pb-20 md:pb-6">
      <div className="flex items-center mb-6">
        <button
          onClick={() => navigate(-1)}
          className="mr-4 p-2 hover:bg-gray-100 rounded-full"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <h1 className="text-2xl font-bold">ポイント送信</h1>
      </div>

      {success ? (
        <div className="bg-white rounded-xl shadow p-8 text-center space-y-4">
          <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto">
            <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 className="text-xl font-bold text-gray-900">送信完了!</h2>
          <p className="text-gray-600">ダッシュボードに戻ります...</p>
        </div>
      ) : (
        <form onSubmit={handleTransfer} className="space-y-4">
          {/* 友達から選択 */}
          <div className="bg-white rounded-xl shadow p-4">
            <label className="block text-sm font-medium text-gray-700 mb-3">
              送信先を選択
            </label>

            {/* 選択済みの友達 */}
            {selectedFriendName && (
              <div className="flex items-center justify-between p-3 bg-primary-50 rounded-lg mb-3">
                <div className="flex items-center">
                  <div className="w-8 h-8 bg-primary-100 rounded-full flex items-center justify-center mr-3">
                    <span className="text-primary-600 font-medium text-sm">
                      {selectedFriendName.charAt(0)}
                    </span>
                  </div>
                  <span className="font-medium text-primary-700">{selectedFriendName}</span>
                </div>
                <button
                  type="button"
                  onClick={() => {
                    setToUserId('');
                    setSelectedFriendName('');
                  }}
                  className="text-primary-400 hover:text-primary-600"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            )}

            {/* 友達リスト */}
            {!selectedFriendName && (
              <>
                {friendsLoading ? (
                  <div className="flex justify-center py-4">
                    <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary-600"></div>
                  </div>
                ) : friends.length > 0 ? (
                  <div className="space-y-1 max-h-48 overflow-y-auto">
                    {friends.map((item) => (
                      <button
                        key={item.friendship.id}
                        type="button"
                        onClick={() => handleSelectFriend(item)}
                        className="w-full flex items-center p-3 rounded-lg hover:bg-gray-50 transition-colors text-left"
                      >
                        <div className="w-8 h-8 bg-gray-100 rounded-full flex items-center justify-center mr-3">
                          <span className="text-gray-600 font-medium text-sm">
                            {item.friend?.display_name?.charAt(0) || '?'}
                          </span>
                        </div>
                        <div>
                          <div className="text-sm font-medium text-gray-900">
                            {item.friend?.display_name}
                          </div>
                          <div className="text-xs text-gray-500">
                            @{item.friend?.username}
                          </div>
                        </div>
                      </button>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-gray-500 text-center py-2">友達がいません</p>
                )}

                {/* 手動入力フォールバック */}
                <div className="mt-3 pt-3 border-t border-gray-200">
                  <label className="block text-xs text-gray-500 mb-1">
                    またはユーザーIDを直接入力
                  </label>
                  <input
                    type="text"
                    value={toUserId}
                    onChange={(e) => setToUserId(e.target.value)}
                    placeholder="ユーザーIDを入力"
                    className="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  />
                </div>
              </>
            )}
          </div>

          {/* 金額 */}
          <div className="bg-white rounded-xl shadow p-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              金額 <span className="text-red-500">*</span>
            </label>
            <div className="relative">
              <input
                type="number"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="100"
                required
                min="1"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              />
              <span className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-500">P</span>
            </div>
          </div>

          {/* メモ */}
          <div className="bg-white rounded-xl shadow p-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              メモ (オプション)
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="メッセージを入力"
              rows={3}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            />
          </div>

          {error && (
            <div className="rounded-lg bg-red-50 p-4">
              <div className="text-sm text-red-800">{error}</div>
            </div>
          )}

          <button
            type="submit"
            disabled={loading || !toUserId}
            className="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? '送信中...' : 'ポイントを送信'}
          </button>
        </form>
      )}
    </div>
  );
};
