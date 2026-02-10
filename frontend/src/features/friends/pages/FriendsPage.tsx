import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { FriendshipRepository } from '@/infrastructure/api/repositories/FriendshipRepository';
import { FriendInfo, PendingRequestInfo } from '@/core/domain/Friendship';
import { axiosInstance } from '@/infrastructure/api/client';

const friendshipRepository = new FriendshipRepository();

export const FriendsPage: React.FC = () => {
  const [friends, setFriends] = useState<FriendInfo[]>([]);
  const [pendingRequests, setPendingRequests] = useState<PendingRequestInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<'friends' | 'pending' | 'add'>('friends');
  const navigate = useNavigate();

  // フレンド申請用state
  const [searchUsername, setSearchUsername] = useState('');
  const [searchResult, setSearchResult] = useState<{ id: string; username: string; display_name: string } | null>(null);
  const [searchError, setSearchError] = useState('');
  const [searching, setSearching] = useState(false);
  const [sendingRequest, setSendingRequest] = useState(false);
  const [requestSuccess, setRequestSuccess] = useState('');

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [friendsData, pendingData] = await Promise.all([
        friendshipRepository.getFriends(),
        friendshipRepository.getPendingRequests(),
      ]);
      setFriends(friendsData.friends || []);
      setPendingRequests(pendingData.requests || []);
    } catch (error) {
      console.error('Failed to load friends:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleAccept = async (friendshipId: string) => {
    try {
      await friendshipRepository.acceptRequest({ friendship_id: friendshipId });
      loadData();
    } catch (error) {
      console.error('Failed to accept request:', error);
    }
  };

  const handleReject = async (friendshipId: string) => {
    try {
      await friendshipRepository.rejectRequest({ friendship_id: friendshipId });
      loadData();
    } catch (error) {
      console.error('Failed to reject request:', error);
    }
  };

  const handleSearch = async () => {
    if (!searchUsername.trim()) return;

    setSearching(true);
    setSearchError('');
    setSearchResult(null);
    setRequestSuccess('');

    try {
      const response = await axiosInstance.get(`/api/users/search?username=${encodeURIComponent(searchUsername.trim())}`);
      setSearchResult(response.data.user);
    } catch (error: any) {
      setSearchError(error.response?.data?.error || 'ユーザーが見つかりません');
    } finally {
      setSearching(false);
    }
  };

  const handleSendRequest = async () => {
    if (!searchResult) return;

    setSendingRequest(true);
    setSearchError('');

    try {
      await friendshipRepository.sendRequest({ addressee_id: searchResult.id });
      setRequestSuccess(`${searchResult.display_name} さんにフレンド申請を送信しました`);
      setSearchResult(null);
      setSearchUsername('');
      loadData();
    } catch (error: any) {
      setSearchError(error.response?.data?.error || 'フレンド申請に失敗しました');
    } finally {
      setSendingRequest(false);
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
        <h1 className="text-2xl font-bold">友達</h1>
      </div>

      {/* タブ */}
      <div className="flex space-x-1 bg-gray-100 p-1 rounded-lg">
        <button
          onClick={() => setActiveTab('friends')}
          className={`flex-1 py-2 px-3 rounded-md text-sm font-medium transition-colors ${activeTab === 'friends'
              ? 'bg-white text-gray-900 shadow'
              : 'text-gray-600 hover:text-gray-900'
            }`}
        >
          友達 ({friends.length})
        </button>
        <button
          onClick={() => setActiveTab('pending')}
          className={`flex-1 py-2 px-3 rounded-md text-sm font-medium transition-colors ${activeTab === 'pending'
              ? 'bg-white text-gray-900 shadow'
              : 'text-gray-600 hover:text-gray-900'
            }`}
        >
          申請 ({pendingRequests.length})
        </button>
        <button
          onClick={() => setActiveTab('add')}
          className={`flex-1 py-2 px-3 rounded-md text-sm font-medium transition-colors ${activeTab === 'add'
              ? 'bg-white text-gray-900 shadow'
              : 'text-gray-600 hover:text-gray-900'
            }`}
        >
          追加
        </button>
      </div>

      {loading ? (
        <div className="flex justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
        </div>
      ) : activeTab === 'friends' ? (
        friends.length === 0 ? (
          <div className="bg-white rounded-xl shadow p-12 text-center">
            <svg className="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
            <p className="text-gray-500 mb-4">まだ友達がいません</p>
            <button
              onClick={() => setActiveTab('add')}
              className="px-4 py-2 bg-primary-600 text-white text-sm rounded-lg hover:bg-primary-700"
            >
              友達を追加する
            </button>
          </div>
        ) : (
          <div className="bg-white rounded-xl shadow divide-y divide-gray-200">
            {friends.map((item) => (
              <div key={item.friendship.id} className="p-4 hover:bg-gray-50">
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    <div className="w-10 h-10 bg-primary-100 rounded-full flex items-center justify-center mr-3">
                      <span className="text-primary-600 font-medium">
                        {item.friend?.display_name?.charAt(0) || '?'}
                      </span>
                    </div>
                    <div>
                      <div className="font-medium text-gray-900">
                        {item.friend?.display_name}
                      </div>
                      <div className="text-sm text-gray-500">
                        @{item.friend?.username}
                      </div>
                    </div>
                  </div>
                  <button
                    onClick={() => navigate('/transfer', {
                      state: {
                        friendId: item.friend?.id,
                        friendName: item.friend?.display_name,
                        friendUsername: item.friend?.username,
                      }
                    })}
                    className="flex items-center px-3 py-1.5 bg-primary-50 text-primary-600 text-sm font-medium rounded-lg hover:bg-primary-100 transition-colors"
                  >
                    <svg className="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                    </svg>
                    送金
                  </button>
                </div>
              </div>
            ))}
          </div>
        )
      ) : activeTab === 'pending' ? (
        pendingRequests.length === 0 ? (
          <div className="bg-white rounded-xl shadow p-12 text-center">
            <svg className="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <p className="text-gray-500">保留中の友達申請はありません</p>
          </div>
        ) : (
          <div className="bg-white rounded-xl shadow divide-y divide-gray-200">
            {pendingRequests.map((item) => (
              <div key={item.friendship.id} className="p-4 hover:bg-gray-50">
                <div className="flex items-center justify-between">
                  <div className="flex items-center flex-1">
                    <div className="w-10 h-10 bg-orange-100 rounded-full flex items-center justify-center mr-3">
                      <span className="text-orange-600 font-medium">
                        {item.requester?.display_name?.charAt(0) || '?'}
                      </span>
                    </div>
                    <div>
                      <div className="font-medium text-gray-900">
                        {item.requester?.display_name}
                      </div>
                      <div className="text-sm text-gray-500">
                        @{item.requester?.username}
                      </div>
                      <div className="text-xs text-gray-400 mt-0.5">
                        {new Date(item.friendship.created_at).toLocaleDateString('ja-JP')}
                      </div>
                    </div>
                  </div>
                  <div className="flex space-x-2">
                    <button
                      onClick={() => handleAccept(item.friendship.id)}
                      className="px-4 py-2 bg-primary-600 text-white text-sm rounded-lg hover:bg-primary-700"
                    >
                      承認
                    </button>
                    <button
                      onClick={() => handleReject(item.friendship.id)}
                      className="px-4 py-2 bg-gray-200 text-gray-700 text-sm rounded-lg hover:bg-gray-300"
                    >
                      拒否
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )
      ) : (
        /* 友達追加タブ */
        <div className="bg-white rounded-xl shadow p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              ユーザー名で検索
            </label>
            <div className="flex space-x-2">
              <div className="relative flex-1">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">@</span>
                <input
                  type="text"
                  value={searchUsername}
                  onChange={(e) => setSearchUsername(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                  placeholder="ユーザー名を入力"
                  className="w-full pl-8 pr-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                />
              </div>
              <button
                onClick={handleSearch}
                disabled={searching || !searchUsername.trim()}
                className="px-4 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {searching ? (
                  <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white"></div>
                ) : (
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                  </svg>
                )}
              </button>
            </div>
          </div>

          {/* 検索結果 */}
          {searchResult && (
            <div className="border border-gray-200 rounded-lg p-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <div className="w-12 h-12 bg-primary-100 rounded-full flex items-center justify-center mr-4">
                    <span className="text-primary-600 font-bold text-lg">
                      {searchResult.display_name.charAt(0)}
                    </span>
                  </div>
                  <div>
                    <div className="font-medium text-gray-900">{searchResult.display_name}</div>
                    <div className="text-sm text-gray-500">@{searchResult.username}</div>
                  </div>
                </div>
                <button
                  onClick={handleSendRequest}
                  disabled={sendingRequest}
                  className="px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 disabled:opacity-50"
                >
                  {sendingRequest ? '送信中...' : '申請する'}
                </button>
              </div>
            </div>
          )}

          {/* エラー */}
          {searchError && (
            <div className="rounded-lg bg-red-50 p-4">
              <div className="text-sm text-red-800">{searchError}</div>
            </div>
          )}

          {/* 成功メッセージ */}
          {requestSuccess && (
            <div className="rounded-lg bg-green-50 p-4">
              <div className="flex items-center">
                <svg className="w-5 h-5 text-green-600 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
                <div className="text-sm text-green-800">{requestSuccess}</div>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
