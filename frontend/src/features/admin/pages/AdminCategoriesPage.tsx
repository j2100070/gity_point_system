import React, { useEffect, useState } from 'react';
import {
    getCategories,
    createCategory,
    updateCategory,
    deleteCategory,
} from '../api/categoriesApi';
import type { Category, CreateCategoryRequest, UpdateCategoryRequest } from '../types/category';

export const AdminCategoriesPage: React.FC = () => {
    const [categories, setCategories] = useState<Category[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [editingCategory, setEditingCategory] = useState<Category | null>(null);

    useEffect(() => {
        loadCategories();
    }, []);

    const loadCategories = async () => {
        try {
            setLoading(true);
            setError(null);
            const data = await getCategories(false); // 管理者は全て表示
            setCategories(data.categories || []);
        } catch (err: any) {
            setError(err.response?.data?.error || 'カテゴリの読み込みに失敗しました');
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async (categoryId: string) => {
        if (!confirm('このカテゴリを削除しますか？')) return;

        try {
            await deleteCategory(categoryId);
            alert('削除しました');
            loadCategories();
        } catch (err: any) {
            alert(err.response?.data?.error || '削除に失敗しました');
        }
    };

    return (
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <div className="mb-8">
                <h1 className="text-3xl font-bold text-gray-900 mb-4">カテゴリ管理</h1>
                <p className="text-gray-600">商品カテゴリの追加・編集・削除ができます</p>
            </div>

            {error && (
                <div className="mb-6 p-4 bg-red-50 border border-red-200 text-red-600 rounded-md">
                    {error}
                </div>
            )}

            <div className="mb-6">
                <button
                    onClick={() => setShowCreateModal(true)}
                    className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
                >
                    + 新しいカテゴリを追加
                </button>
            </div>

            {loading ? (
                <div className="text-center py-12">読み込み中...</div>
            ) : (
                <div className="bg-white shadow-md rounded-lg overflow-hidden">
                    <table className="min-w-full divide-y divide-gray-200">
                        <thead className="bg-gray-50">
                            <tr>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    カテゴリ名
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    コード
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    説明
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    表示順
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    状態
                                </th>
                                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    操作
                                </th>
                            </tr>
                        </thead>
                        <tbody className="bg-white divide-y divide-gray-200">
                            {categories.map((category) => (
                                <tr key={category.ID}>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <div className="text-sm font-medium text-gray-900">{category.Name}</div>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <code className="px-2 py-1 bg-gray-100 rounded text-sm">{category.Code}</code>
                                    </td>
                                    <td className="px-6 py-4">
                                        <div className="text-sm text-gray-500">{category.Description || '-'}</div>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                                        {category.DisplayOrder}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <span
                                            className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${category.IsActive
                                                    ? 'bg-green-100 text-green-800'
                                                    : 'bg-red-100 text-red-800'
                                                }`}
                                        >
                                            {category.IsActive ? '有効' : '無効'}
                                        </span>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                        <button
                                            onClick={() => setEditingCategory(category)}
                                            className="text-blue-600 hover:text-blue-900 mr-4"
                                        >
                                            編集
                                        </button>
                                        <button
                                            onClick={() => handleDelete(category.ID)}
                                            className="text-red-600 hover:text-red-900"
                                        >
                                            削除
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}

            {/* カテゴリ作成/編集モーダル */}
            {(showCreateModal || editingCategory) && (
                <CategoryFormModal
                    category={editingCategory}
                    onClose={() => {
                        setShowCreateModal(false);
                        setEditingCategory(null);
                    }}
                    onSuccess={() => {
                        setShowCreateModal(false);
                        setEditingCategory(null);
                        loadCategories();
                    }}
                />
            )}
        </div>
    );
};

// カテゴリフォームモーダルコンポーネント
const CategoryFormModal: React.FC<{
    category: Category | null;
    onClose: () => void;
    onSuccess: () => void;
}> = ({ category, onClose, onSuccess }) => {
    const [formData, setFormData] = useState<CreateCategoryRequest & { is_active?: boolean }>({
        name: category?.Name || '',
        code: category?.Code || '',
        description: category?.Description || '',
        display_order: category?.DisplayOrder || 0,
        is_active: category?.IsActive ?? true,
    });
    const [submitting, setSubmitting] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        try {
            setSubmitting(true);
            setError(null);

            if (category) {
                const updateReq: UpdateCategoryRequest = {
                    name: formData.name,
                    description: formData.description,
                    display_order: formData.display_order,
                    is_active: formData.is_active!,
                };
                await updateCategory(category.ID, updateReq);
            } else {
                await createCategory(formData);
            }

            alert(category ? '更新しました' : '作成しました');
            onSuccess();
        } catch (err: any) {
            setError(err.response?.data?.error || '処理に失敗しました');
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-8 max-w-lg w-full max-h-[90vh] overflow-y-auto">
                <h2 className="text-2xl font-bold mb-6">
                    {category ? 'カテゴリを編集' : '新しいカテゴリを作成'}
                </h2>

                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">カテゴリ名</label>
                        <input
                            type="text"
                            value={formData.name}
                            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                            required
                            className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
                            placeholder="例: 飲み物"
                        />
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            コード {category && <span className="text-gray-400">(変更不可)</span>}
                        </label>
                        <input
                            type="text"
                            value={formData.code}
                            onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                            required
                            disabled={!!category}
                            className={`w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 ${category ? 'bg-gray-100 cursor-not-allowed' : ''
                                }`}
                            placeholder="例: drink"
                        />
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">説明</label>
                        <textarea
                            value={formData.description || ''}
                            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                            rows={2}
                            className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
                            placeholder="カテゴリの説明（任意）"
                        />
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">表示順序</label>
                        <input
                            type="number"
                            value={formData.display_order}
                            onChange={(e) =>
                                setFormData({ ...formData, display_order: parseInt(e.target.value) || 0 })
                            }
                            min="0"
                            className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
                        />
                    </div>

                    {category && (
                        <div className="flex items-center">
                            <input
                                type="checkbox"
                                id="is_active"
                                checked={formData.is_active}
                                onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                                className="mr-2"
                            />
                            <label htmlFor="is_active" className="text-sm font-medium text-gray-700">
                                有効
                            </label>
                        </div>
                    )}

                    {error && (
                        <div className="p-4 bg-red-50 border border-red-200 text-red-600 rounded-md text-sm">
                            {error}
                        </div>
                    )}

                    <div className="flex justify-end space-x-4">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
                        >
                            キャンセル
                        </button>
                        <button
                            type="submit"
                            disabled={submitting}
                            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:bg-gray-300"
                        >
                            {submitting ? '処理中...' : category ? '更新' : '作成'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};
