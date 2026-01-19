import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AdminRepository } from '@/infrastructure/api/repositories/AdminRepository';
import { User } from '@/core/domain/User';

const adminRepository = new AdminRepository();

export const AdminUsersPage: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [modalType, setModalType] = useState<'grant' | 'deduct' | 'role' | null>(null);
  const [amount, setAmount] = useState('');
  const [description, setDescription] = useState('');
  const [newRole, setNewRole] = useState<'user' | 'admin'>('user');
  const [processing, setProcessing] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    try {
      const data = await adminRepository.getAllUsers(0, 100);
      setUsers(data.users);
    } catch (error) {
      console.error('Failed to load users:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleGrantPoints = async () => {
    if (!selectedUser || !amount) return;
    setProcessing(true);
    try {
      await adminRepository.grantPoints(
        selectedUser.id,
        parseInt(amount),
        description || undefined
      );
      alert('ポイント付与が完了しました');
      closeModal();
      loadUsers();
    } catch (error: any) {
      alert(error.response?.data?.error || 'ポイント付与に失敗しました');
    } finally {
      setProcessing(false);
    }
  };

  const handleDeductPoints = async () => {
    if (!selectedUser || !amount) return;
    setProcessing(true);
    try {
      await adminRepository.deductPoints(
        selectedUser.id,
        parseInt(amount),
        description || undefined
      );
      alert('ポイント減算が完了しました');
      closeModal();
      loadUsers();
    } catch (error: any) {
      alert(error.response?.data?.error || 'ポイント減算に失敗しました');
    } finally {
      setProcessing(false);
    }
  };

  const handleChangeRole = async () => {
    if (!selectedUser) return;
    setProcessing(true);
    try {
      await adminRepository.changeUserRole(selectedUser.id, newRole);
      alert('役割変更が完了しました');
      closeModal();
      loadUsers();
    } catch (error: any) {
      alert(error.response?.data?.error || '役割変更に失敗しました');
    } finally {
      setProcessing(false);
    }
  };

  const handleDeactivateUser = async (userId: string) => {
    if (!confirm('本当にこのユーザーを無効化しますか？')) return;
    try {
      await adminRepository.deactivateUser(userId);
      alert('ユーザーを無効化しました');
      loadUsers();
    } catch (error: any) {
      alert(error.response?.data?.error || 'ユーザー無効化に失敗しました');
    }
  };

  const openModal = (user: User, type: 'grant' | 'deduct' | 'role') => {
    setSelectedUser(user);
    setModalType(type);
    setAmount('');
    setDescription('');
    setNewRole(user.role);
  };

  const closeModal = () => {
    setSelectedUser(null);
    setModalType(null);
    setAmount('');
    setDescription('');
  };

  return (
    <div className="max-w-7xl mx-auto space-y-6 pb-20 md:pb-6">
      <div className="flex items-center mb-6">
        <button
          onClick={() => navigate(-1)}
          className="mr-4 p-2 hover:bg-gray-100 rounded-full"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <h1 className="text-2xl font-bold">ユーザー管理</h1>
      </div>

      {loading ? (
        <div className="flex justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
        </div>
      ) : (
        <div className="bg-white rounded-xl shadow overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    ユーザー
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    メール
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    残高
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    役割
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    ステータス
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {users.map((user) => (
                  <tr key={user.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <div>
                          <div className="text-sm font-medium text-gray-900">{user.display_name}</div>
                          <div className="text-sm text-gray-500">@{user.username}</div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm text-gray-900">{user.email}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900">
                        {user.balance.toLocaleString()} P
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${
                        user.role === 'admin'
                          ? 'bg-red-100 text-red-800'
                          : 'bg-green-100 text-green-800'
                      }`}>
                        {user.role === 'admin' ? '管理者' : 'ユーザー'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${
                        user.is_active
                          ? 'bg-green-100 text-green-800'
                          : 'bg-gray-100 text-gray-800'
                      }`}>
                        {user.is_active ? '有効' : '無効'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
                      <button
                        onClick={() => openModal(user, 'grant')}
                        className="text-green-600 hover:text-green-900"
                      >
                        付与
                      </button>
                      <button
                        onClick={() => openModal(user, 'deduct')}
                        className="text-orange-600 hover:text-orange-900"
                      >
                        減算
                      </button>
                      <button
                        onClick={() => openModal(user, 'role')}
                        className="text-blue-600 hover:text-blue-900"
                      >
                        役割
                      </button>
                      {user.is_active && (
                        <button
                          onClick={() => handleDeactivateUser(user.id)}
                          className="text-red-600 hover:text-red-900"
                        >
                          無効化
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* モーダル */}
      {modalType && selectedUser && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl max-w-md w-full p-6 space-y-4">
            <h2 className="text-xl font-bold">
              {modalType === 'grant' && 'ポイント付与'}
              {modalType === 'deduct' && 'ポイント減算'}
              {modalType === 'role' && '役割変更'}
            </h2>

            <div className="bg-gray-50 p-4 rounded-lg">
              <div className="text-sm text-gray-600">対象ユーザー</div>
              <div className="font-medium">{selectedUser.display_name}</div>
              <div className="text-sm text-gray-500">現在の残高: {selectedUser.balance.toLocaleString()} P</div>
            </div>

            {modalType !== 'role' ? (
              <>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    金額
                  </label>
                  <div className="relative">
                    <input
                      type="number"
                      value={amount}
                      onChange={(e) => setAmount(e.target.value)}
                      className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                      placeholder="1000"
                    />
                    <span className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-500">P</span>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    理由（オプション）
                  </label>
                  <textarea
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    rows={3}
                    placeholder="理由を入力"
                  />
                </div>
              </>
            ) : (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  新しい役割
                </label>
                <select
                  value={newRole}
                  onChange={(e) => setNewRole(e.target.value as 'user' | 'admin')}
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                >
                  <option value="user">ユーザー</option>
                  <option value="admin">管理者</option>
                </select>
              </div>
            )}

            <div className="flex space-x-3">
              <button
                onClick={closeModal}
                disabled={processing}
                className="flex-1 px-4 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 disabled:opacity-50"
              >
                キャンセル
              </button>
              <button
                onClick={
                  modalType === 'grant'
                    ? handleGrantPoints
                    : modalType === 'deduct'
                    ? handleDeductPoints
                    : handleChangeRole
                }
                disabled={processing || (modalType !== 'role' && !amount)}
                className="flex-1 px-4 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {processing ? '処理中...' : '実行'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
