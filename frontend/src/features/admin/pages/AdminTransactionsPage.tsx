import React, { useEffect, useState, useCallback } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { AdminRepository } from '@/infrastructure/api/repositories/AdminRepository';
import { Transaction } from '@/core/domain/Transaction';

const adminRepository = new AdminRepository();
const PAGE_SIZE = 20;

type SortField = '' | 'created_at' | 'amount';
type SortOrder = 'asc' | 'desc';

export const AdminTransactionsPage: React.FC = () => {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [totalCount, setTotalCount] = useState(0);
  const [currentPage, setCurrentPage] = useState(0);
  const [searchParams, setSearchParams] = useSearchParams();

  // URLからフィルタ初期値を取得
  const [filter, setFilter] = useState<string>(searchParams.get('type') || 'all');
  const [dateFrom, setDateFrom] = useState(searchParams.get('date_from') || '');
  const [dateTo, setDateTo] = useState(searchParams.get('date_to') || '');
  const [sortBy, setSortBy] = useState<SortField>((searchParams.get('sort_by') as SortField) || '');
  const [sortOrder, setSortOrder] = useState<SortOrder>((searchParams.get('sort_order') as SortOrder) || 'desc');
  const navigate = useNavigate();

  useEffect(() => {
    loadTransactions();
  }, [currentPage, filter, dateFrom, dateTo, sortBy, sortOrder]);

  const loadTransactions = async () => {
    setLoading(true);
    try {
      const data = await adminRepository.getAllTransactions(
        currentPage * PAGE_SIZE,
        PAGE_SIZE,
        filter !== 'all' ? filter : undefined,
        dateFrom || undefined,
        dateTo || undefined,
        sortBy || undefined,
        sortBy ? sortOrder : undefined,
      );
      setTransactions(data.transactions);
      setTotalCount(data.total);
    } catch (error) {
      console.error('Failed to load transactions:', error);
    } finally {
      setLoading(false);
    }
  };

  const totalPages = Math.ceil(totalCount / PAGE_SIZE);

  const handleFilterChange = useCallback((newFilter: string) => {
    setFilter(newFilter);
    setCurrentPage(0);
    // URLにフィルタを反映
    const params = new URLSearchParams();
    if (newFilter !== 'all') params.set('type', newFilter);
    if (dateFrom) params.set('date_from', dateFrom);
    if (dateTo) params.set('date_to', dateTo);
    setSearchParams(params, { replace: true });
  }, [dateFrom, dateTo, setSearchParams]);

  const handleDateChange = useCallback((field: 'from' | 'to', value: string) => {
    if (field === 'from') setDateFrom(value);
    else setDateTo(value);
    setCurrentPage(0);
  }, []);

  const handleSort = useCallback((field: SortField) => {
    if (sortBy === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortBy(field);
      setSortOrder(field === 'amount' ? 'desc' : 'desc');
    }
    setCurrentPage(0);
  }, [sortBy, sortOrder]);

  const renderSortIcon = (field: SortField) => {
    if (sortBy !== field) return <span className="ml-1 text-gray-300">⇅</span>;
    return <span className="ml-1">{sortOrder === 'asc' ? '▲' : '▼'}</span>;
  };

  const clearFilters = () => {
    setFilter('all');
    setDateFrom('');
    setDateTo('');
    setSortBy('');
    setSortOrder('desc');
    setCurrentPage(0);
    setSearchParams({}, { replace: true });
  };

  const hasActiveFilters = filter !== 'all' || dateFrom || dateTo;

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

  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'transfer':
        return { label: 'ユーザー送信', color: 'bg-blue-100 text-blue-800' };
      case 'admin_grant':
        return { label: '管理者付与', color: 'bg-green-100 text-green-800' };
      case 'admin_deduct':
        return { label: '管理者減算', color: 'bg-orange-100 text-orange-800' };
      case 'system_grant':
        return { label: 'システム付与', color: 'bg-purple-100 text-purple-800' };
      case 'daily_bonus':
        return { label: '出勤ボーナス', color: 'bg-teal-100 text-teal-800' };
      case 'system_expire':
        return { label: 'システム失効', color: 'bg-red-100 text-red-800' };
      default:
        return { label: type, color: 'bg-gray-100 text-gray-800' };
    }
  };

  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'completed':
        return { label: '完了', color: 'bg-green-100 text-green-800' };
      case 'pending':
        return { label: '保留', color: 'bg-yellow-100 text-yellow-800' };
      case 'failed':
        return { label: '失敗', color: 'bg-red-100 text-red-800' };
      case 'reversed':
        return { label: '取消', color: 'bg-gray-100 text-gray-800' };
      default:
        return { label: status, color: 'bg-gray-100 text-gray-800' };
    }
  };

  // ページネーションで表示するページ番号の範囲を計算
  const getPageNumbers = () => {
    const pages: number[] = [];
    const maxVisible = 5;
    let start = Math.max(0, currentPage - Math.floor(maxVisible / 2));
    let end = Math.min(totalPages - 1, start + maxVisible - 1);
    start = Math.max(0, end - maxVisible + 1);
    for (let i = start; i <= end; i++) {
      pages.push(i);
    }
    return pages;
  };

  const filterButtons = [
    { key: 'all', label: '全て' },
    { key: 'transfer', label: 'ユーザー送信' },
    { key: 'admin_grant', label: '管理者付与' },
    { key: 'admin_deduct', label: '管理者減算' },
    { key: 'system_grant', label: 'システム付与' },
    { key: 'daily_bonus', label: '出勤ボーナス' },
    { key: 'system_expire', label: 'システム失効' },
  ];

  return (
    <div className="max-w-7xl mx-auto space-y-6 pb-20 md:pb-6">
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
          <h1 className="text-2xl font-bold">全トランザクション</h1>
        </div>
        <div className="text-sm text-gray-500">
          全 {totalCount.toLocaleString()} 件
        </div>
      </div>

      {/* フィルター */}
      <div className="bg-white rounded-xl shadow p-4 space-y-4">
        <div className="flex flex-wrap gap-2">
          {filterButtons.map(({ key, label }) => (
            <button
              key={key}
              onClick={() => handleFilterChange(key)}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${filter === key
                ? 'bg-primary-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
            >
              {label}
            </button>
          ))}
        </div>

        {/* 日付フィルター */}
        <div className="flex flex-wrap items-center gap-3">
          <label className="text-sm text-gray-500 font-medium">期間:</label>
          <input
            type="date"
            value={dateFrom}
            onChange={(e) => handleDateChange('from', e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
          <span className="text-gray-400">〜</span>
          <input
            type="date"
            value={dateTo}
            onChange={(e) => handleDateChange('to', e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
          {hasActiveFilters && (
            <button
              onClick={clearFilters}
              className="px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded-lg transition-colors"
            >
              フィルタをクリア
            </button>
          )}
        </div>
      </div>

      {loading ? (
        <div className="flex justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
        </div>
      ) : (
        <>
          <div className="bg-white rounded-xl shadow overflow-hidden">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 select-none"
                      onClick={() => handleSort('created_at')}
                    >
                      日時{renderSortIcon('created_at')}
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      種類
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      送信元
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      送信先
                    </th>
                    <th
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 select-none"
                      onClick={() => handleSort('amount')}
                    >
                      金額{renderSortIcon('amount')}
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      ステータス
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      説明
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {transactions.map((tx) => {
                    const typeInfo = getTypeLabel(tx.transaction_type);
                    const statusInfo = getStatusLabel(tx.status);
                    return (
                      <tr key={tx.id} className="hover:bg-gray-50">
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {formatDate(tx.created_at)}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${typeInfo.color}`}>
                            {typeInfo.label}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {tx.from_user ? (
                            <div>
                              <div className="font-medium">{tx.from_user.display_name}</div>
                              <div className="text-gray-500">@{tx.from_user.username}</div>
                            </div>
                          ) : (
                            <span className="text-gray-400">-</span>
                          )}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {tx.to_user ? (
                            <div>
                              <div className="font-medium">{tx.to_user.display_name}</div>
                              <div className="text-gray-500">@{tx.to_user.username}</div>
                            </div>
                          ) : (
                            <span className="text-gray-400">-</span>
                          )}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                          {tx.amount.toLocaleString()} P
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${statusInfo.color}`}>
                            {statusInfo.label}
                          </span>
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-500 max-w-xs truncate">
                          {tx.description || '-'}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
            {transactions.length === 0 && (
              <div className="text-center py-12">
                <svg className="w-16 h-16 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p className="text-gray-500">該当するトランザクションがありません</p>
              </div>
            )}
          </div>

          {/* ページネーション */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between bg-white rounded-xl shadow px-6 py-4">
              <div className="text-sm text-gray-500">
                {currentPage * PAGE_SIZE + 1}〜{Math.min((currentPage + 1) * PAGE_SIZE, totalCount)} 件 / {totalCount.toLocaleString()} 件
              </div>
              <div className="flex items-center space-x-1">
                <button
                  onClick={() => setCurrentPage(0)}
                  disabled={currentPage === 0}
                  className="px-3 py-2 text-sm rounded-lg border border-gray-300 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  «
                </button>
                <button
                  onClick={() => setCurrentPage(currentPage - 1)}
                  disabled={currentPage === 0}
                  className="px-3 py-2 text-sm rounded-lg border border-gray-300 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  ‹
                </button>
                {getPageNumbers().map((page) => (
                  <button
                    key={page}
                    onClick={() => setCurrentPage(page)}
                    className={`px-3 py-2 text-sm rounded-lg border ${page === currentPage
                      ? 'bg-primary-600 text-white border-primary-600'
                      : 'border-gray-300 hover:bg-gray-50'
                      }`}
                  >
                    {page + 1}
                  </button>
                ))}
                <button
                  onClick={() => setCurrentPage(currentPage + 1)}
                  disabled={currentPage >= totalPages - 1}
                  className="px-3 py-2 text-sm rounded-lg border border-gray-300 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  ›
                </button>
                <button
                  onClick={() => setCurrentPage(totalPages - 1)}
                  disabled={currentPage >= totalPages - 1}
                  className="px-3 py-2 text-sm rounded-lg border border-gray-300 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  »
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
};
