import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { TransferRequestRepository } from '@/infrastructure/api/repositories/TransferRequestRepository';
import { TransferRequestInfo } from '@/core/domain/TransferRequest';

const transferRequestRepository = new TransferRequestRepository();

export const TransferRequestsPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'received' | 'sent'>('received');
  const [receivedRequests, setReceivedRequests] = useState<TransferRequestInfo[]>([]);
  const [sentRequests, setSentRequests] = useState<TransferRequestInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    loadRequests();
  }, [activeTab]);

  const loadRequests = async () => {
    setLoading(true);
    setError('');
    try {
      if (activeTab === 'received') {
        const response = await transferRequestRepository.getPendingRequests();
        setReceivedRequests(response.requests || []);
      } else {
        const response = await transferRequestRepository.getSentRequests();
        setSentRequests(response.requests || []);
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'リクエストの取得に失敗しました');
    } finally {
      setLoading(false);
    }
  };

  const handleApprove = async (requestId: string) => {
    if (!window.confirm('この送金リクエストを承認しますか？')) return;

    try {
      await transferRequestRepository.approveTransferRequest(requestId);
      loadRequests();
    } catch (err: any) {
      alert(err.response?.data?.error || '承認に失敗しました');
    }
  };

  const handleReject = async (requestId: string) => {
    if (!window.confirm('この送金リクエストを拒否しますか？')) return;

    try {
      await transferRequestRepository.rejectTransferRequest(requestId);
      loadRequests();
    } catch (err: any) {
      alert(err.response?.data?.error || '拒否に失敗しました');
    }
  };

  const handleCancel = async (requestId: string) => {
    if (!window.confirm('この送金リクエストをキャンセルしますか？')) return;

    try {
      await transferRequestRepository.cancelTransferRequest(requestId);
      loadRequests();
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

  const getStatusBadge = (status: string) => {
    const badges = {
      pending: { text: '承認待ち', color: 'bg-yellow-100 text-yellow-800' },
      approved: { text: '承認済み', color: 'bg-green-100 text-green-800' },
      rejected: { text: '拒否', color: 'bg-red-100 text-red-800' },
      cancelled: { text: 'キャンセル', color: 'bg-gray-100 text-gray-800' },
      expired: { text: '期限切れ', color: 'bg-gray-100 text-gray-500' },
    };
    const badge = badges[status as keyof typeof badges] || { text: status, color: 'bg-gray-100' };
    return (
      <span className={`px-2 py-1 text-xs font-medium rounded-full ${badge.color}`}>
        {badge.text}
      </span>
    );
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
        <h1 className="text-2xl font-bold">送金リクエスト</h1>
      </div>

      {/* タブ */}
      <div className="flex space-x-1 bg-gray-100 p-1 rounded-lg">
        <button
          onClick={() => setActiveTab('received')}
          className={`flex-1 py-2 px-3 rounded-md text-sm font-medium transition-colors ${
            activeTab === 'received'
              ? 'bg-white text-gray-900 shadow'
              : 'text-gray-600 hover:text-gray-900'
          }`}
        >
          受信 ({receivedRequests.filter(r => r.transfer_request.status === 'pending').length})
        </button>
        <button
          onClick={() => setActiveTab('sent')}
          className={`flex-1 py-2 px-3 rounded-md text-sm font-medium transition-colors ${
            activeTab === 'sent'
              ? 'bg-white text-gray-900 shadow'
              : 'text-gray-600 hover:text-gray-900'
          }`}
        >
          送信 ({sentRequests.filter(r => r.transfer_request.status === 'pending').length})
        </button>
      </div>

      {loading ? (
        <div className="flex justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
        </div>
      ) : error ? (
        <div className="bg-white rounded-xl shadow p-6">
          <div className="rounded-lg bg-red-50 p-4">
            <div className="text-sm text-red-800">{error}</div>
          </div>
        </div>
      ) : activeTab === 'received' ? (
        receivedRequests.length === 0 ? (
          <div className="bg-white rounded-xl shadow p-12 text-center">
            <svg className="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <p className="text-gray-500">受信した送金リクエストはありません</p>
          </div>
        ) : (
          <div className="bg-white rounded-xl shadow divide-y divide-gray-200">
            {receivedRequests.map((item) => (
              <div key={item.transfer_request.id} className="p-4 hover:bg-gray-50">
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center flex-1">
                    <div className="w-12 h-12 rounded-full overflow-hidden bg-primary-100 flex items-center justify-center mr-3">
                      {item.from_user.avatar_url ? (
                        <img
                          src={item.from_user.avatar_url}
                          alt="Avatar"
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        <span className="text-primary-600 font-bold text-lg">
                          {item.from_user.display_name?.charAt(0) || '?'}
                        </span>
                      )}
                    </div>
                    <div className="flex-1">
                      <div className="font-medium text-gray-900">{item.from_user.display_name}</div>
                      <div className="text-sm text-gray-500">@{item.from_user.username}</div>
                      <div className="text-xs text-gray-400 mt-1">
                        {formatDate(item.transfer_request.created_at)}
                      </div>
                    </div>
                  </div>
                  {getStatusBadge(item.transfer_request.status)}
                </div>

                <div className="mb-3">
                  <div className="text-3xl font-bold text-primary-600 mb-2">
                    {item.transfer_request.amount.toLocaleString()} P
                  </div>
                  {item.transfer_request.message && (
                    <div className="text-sm text-gray-600 bg-gray-50 p-3 rounded-lg">
                      {item.transfer_request.message}
                    </div>
                  )}
                </div>

                {item.transfer_request.status === 'pending' && (
                  <div className="flex space-x-2">
                    <button
                      onClick={() => handleApprove(item.transfer_request.id)}
                      className="flex-1 px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700"
                    >
                      受け取る
                    </button>
                    <button
                      onClick={() => handleReject(item.transfer_request.id)}
                      className="flex-1 px-4 py-2 bg-gray-200 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-300"
                    >
                      拒否
                    </button>
                  </div>
                )}
              </div>
            ))}
          </div>
        )
      ) : (
        sentRequests.length === 0 ? (
          <div className="bg-white rounded-xl shadow p-12 text-center">
            <svg className="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
            </svg>
            <p className="text-gray-500 mb-4">送信した送金リクエストはありません</p>
            <button
              onClick={() => navigate('/qr/scan')}
              className="px-4 py-2 bg-primary-600 text-white text-sm rounded-lg hover:bg-primary-700"
            >
              QRコードをスキャン
            </button>
          </div>
        ) : (
          <div className="bg-white rounded-xl shadow divide-y divide-gray-200">
            {sentRequests.map((item) => (
              <div key={item.transfer_request.id} className="p-4 hover:bg-gray-50">
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center flex-1">
                    <div className="w-12 h-12 rounded-full overflow-hidden bg-green-100 flex items-center justify-center mr-3">
                      {item.to_user.avatar_url ? (
                        <img
                          src={item.to_user.avatar_url}
                          alt="Avatar"
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        <span className="text-green-600 font-bold text-lg">
                          {item.to_user.display_name?.charAt(0) || '?'}
                        </span>
                      )}
                    </div>
                    <div className="flex-1">
                      <div className="font-medium text-gray-900">{item.to_user.display_name}</div>
                      <div className="text-sm text-gray-500">@{item.to_user.username}</div>
                      <div className="text-xs text-gray-400 mt-1">
                        {formatDate(item.transfer_request.created_at)}
                      </div>
                    </div>
                  </div>
                  {getStatusBadge(item.transfer_request.status)}
                </div>

                <div className="mb-3">
                  <div className="text-3xl font-bold text-gray-900 mb-2">
                    {item.transfer_request.amount.toLocaleString()} P
                  </div>
                  {item.transfer_request.message && (
                    <div className="text-sm text-gray-600 bg-gray-50 p-3 rounded-lg">
                      {item.transfer_request.message}
                    </div>
                  )}
                </div>

                {item.transfer_request.status === 'pending' && (
                  <button
                    onClick={() => handleCancel(item.transfer_request.id)}
                    className="w-full px-4 py-2 bg-gray-200 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-300"
                  >
                    キャンセル
                  </button>
                )}
              </div>
            ))}
          </div>
        )
      )}
    </div>
  );
};
