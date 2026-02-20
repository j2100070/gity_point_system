import React, { useEffect, useState } from 'react';
import {
  getProducts,
  createProduct,
  updateProduct,
  deleteProduct,
  getAllExchanges,
  markExchangeDelivered,
} from '../api/productsApi';
import { getCategories } from '@/features/admin/api/categoriesApi';
import type { Category } from '@/features/admin/types/category';
import type { Product, ProductExchange, CreateProductRequest } from '../types';

type Tab = 'products' | 'exchanges';

export const AdminProductsPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<Tab>('products');
  const [products, setProducts] = useState<Product[]>([]);
  const [exchanges, setExchanges] = useState<ProductExchange[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);

  useEffect(() => {
    loadCategories();
  }, []);

  useEffect(() => {
    if (activeTab === 'products') {
      loadProducts();
    } else {
      loadExchanges();
    }
  }, [activeTab]);

  const loadCategories = async () => {
    try {
      const data = await getCategories(true);
      setCategories(data.categories || []);
    } catch (err: any) {
      console.error('Failed to load categories', err);
    }
  };

  const loadProducts = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getProducts({ available_only: false });
      setProducts(data.Products || []);
    } catch (err: any) {
      setError(err.response?.data?.error || '商品の読み込みに失敗しました');
    } finally {
      setLoading(false);
    }
  };

  const loadExchanges = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getAllExchanges();
      setExchanges(data.Exchanges || []);
    } catch (err: any) {
      setError(err.response?.data?.error || '交換履歴の読み込みに失敗しました');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (productId: string) => {
    if (!confirm('この商品を削除しますか？')) return;

    try {
      await deleteProduct(productId);
      alert('削除しました');
      loadProducts();
    } catch (err: any) {
      alert(err.response?.data?.error || '削除に失敗しました');
    }
  };

  const handleMarkDelivered = async (exchangeId: string) => {
    if (!confirm('この交換を配達完了にしますか？')) return;

    try {
      await markExchangeDelivered(exchangeId);
      alert('配達完了にしました');
      loadExchanges();
    } catch (err: any) {
      alert(err.response?.data?.error || '更新に失敗しました');
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

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-4">商品管理</h1>
        <p className="text-gray-600">商品の管理と交換履歴の確認ができます</p>
      </div>

      {/* タブ */}
      <div className="mb-6 border-b border-gray-200">
        <nav className="flex space-x-8">
          <button
            onClick={() => setActiveTab('products')}
            className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${activeTab === 'products'
              ? 'border-blue-500 text-blue-600'
              : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
          >
            商品一覧
          </button>
          <button
            onClick={() => setActiveTab('exchanges')}
            className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${activeTab === 'exchanges'
              ? 'border-blue-500 text-blue-600'
              : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
          >
            交換履歴
          </button>
        </nav>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 text-red-600 rounded-md">
          {error}
        </div>
      )}

      {loading ? (
        <div className="text-center py-12">読み込み中...</div>
      ) : activeTab === 'products' ? (
        <>
          <div className="mb-6">
            <button
              onClick={() => setShowCreateModal(true)}
              className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
            >
              + 新しい商品を追加
            </button>
          </div>

          <div className="bg-white shadow-md rounded-lg overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    商品名
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    カテゴリ
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    価格
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    在庫
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
                {products.map((product) => (
                  <tr key={product.ID}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-gray-900">{product.Name}</div>
                      <div className="text-sm text-gray-500">{product.Description}</div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full bg-blue-100 text-blue-800">
                        {categories.find((c) => c.Code === product.CategoryCode)?.Name || product.CategoryCode}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {product.Price} pt
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {product.Stock === -1 ? '無制限' : product.Stock}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full ${product.IsAvailable
                          ? 'bg-green-100 text-green-800'
                          : 'bg-red-100 text-red-800'
                          }`}
                      >
                        {product.IsAvailable ? '販売中' : '停止中'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => setEditingProduct(product)}
                        className="text-blue-600 hover:text-blue-900 mr-4"
                      >
                        編集
                      </button>
                      <button
                        onClick={() => handleDelete(product.ID)}
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
        </>
      ) : (
        <div className="space-y-4">
          {exchanges.length === 0 ? (
            <div className="text-center py-12 text-gray-500">交換履歴がありません</div>
          ) : (
            exchanges.map((exchange) => (
              <div key={exchange.ID} className="bg-white rounded-lg shadow-md p-6">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                  <div>
                    <div className="text-sm text-gray-600 mb-1">ユーザーID</div>
                    <div className="font-mono text-sm">{exchange.UserID}</div>
                  </div>
                  <div>
                    <div className="text-sm text-gray-600 mb-1">商品ID</div>
                    <div className="font-mono text-sm">{exchange.ProductID}</div>
                  </div>
                  <div>
                    <div className="text-sm text-gray-600 mb-1">数量 / ポイント</div>
                    <div className="font-semibold">
                      {exchange.Quantity}個 / {exchange.PointsUsed}pt
                    </div>
                  </div>
                </div>

                <div className="flex justify-between items-center">
                  <div>
                    <span
                      className={`px-3 py-1 text-sm font-semibold rounded ${exchange.Status === 'delivered'
                        ? 'bg-blue-100 text-blue-800'
                        : exchange.Status === 'completed'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-yellow-100 text-yellow-800'
                        }`}
                    >
                      {exchange.Status}
                    </span>
                    <span className="ml-4 text-sm text-gray-500">
                      {formatDate(exchange.CreatedAt)}
                    </span>
                  </div>

                  {exchange.Status === 'completed' && !exchange.DeliveredAt && (
                    <button
                      onClick={() => handleMarkDelivered(exchange.ID)}
                      className="bg-blue-600 text-white px-4 py-2 rounded-md text-sm hover:bg-blue-700"
                    >
                      配達完了にする
                    </button>
                  )}
                </div>

                {exchange.Notes && (
                  <div className="mt-4 text-sm text-gray-600">
                    備考: {exchange.Notes}
                  </div>
                )}
              </div>
            ))
          )}
        </div>
      )}

      {/* 商品作成/編集モーダル */}
      {(showCreateModal || editingProduct) && (
        <ProductFormModal
          product={editingProduct}
          categories={categories}
          onClose={() => {
            setShowCreateModal(false);
            setEditingProduct(null);
          }}
          onSuccess={() => {
            setShowCreateModal(false);
            setEditingProduct(null);
            loadProducts();
          }}
        />
      )}
    </div>
  );
};

// 商品フォームモーダルコンポーネント
const ProductFormModal: React.FC<{
  product: Product | null;
  categories: Category[];
  onClose: () => void;
  onSuccess: () => void;
}> = ({ product, categories, onClose, onSuccess }) => {
  const [formData, setFormData] = useState<CreateProductRequest & { is_available?: boolean }>({
    name: product?.Name || '',
    description: product?.Description || '',
    category: product?.CategoryCode || 'snack',
    price: product?.Price || 0,
    stock: product?.Stock || 0,
    image_url: product?.ImageURL || '',
    is_available: product?.IsAvailable ?? true,
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      setSubmitting(true);
      setError(null);

      if (product) {
        await updateProduct(product.ID, {
          ...formData,
          is_available: formData.is_available!,
        });
      } else {
        await createProduct(formData);
      }

      alert(product ? '更新しました' : '作成しました');
      onSuccess();
    } catch (err: any) {
      setError(err.response?.data?.error || '処理に失敗しました');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-8 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <h2 className="text-2xl font-bold mb-6">
          {product ? '商品を編集' : '新しい商品を作成'}
        </h2>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">商品名</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              required
              className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">説明</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              required
              rows={3}
              className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">カテゴリ</label>
            <select
              value={formData.category}
              onChange={(e) =>
                setFormData({ ...formData, category: e.target.value })
              }
              className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
            >
              {categories.map((cat) => (
                <option key={cat.Code} value={cat.Code}>
                  {cat.Name}
                </option>
              ))}
            </select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">価格 (pt)</label>
              <input
                type="number"
                value={formData.price}
                onChange={(e) => setFormData({ ...formData, price: parseInt(e.target.value) })}
                required
                min="0"
                className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                在庫 (-1で無制限)
              </label>
              <input
                type="number"
                value={formData.stock}
                onChange={(e) => setFormData({ ...formData, stock: parseInt(e.target.value) })}
                required
                min="-1"
                className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              画像URL（任意）
            </label>
            <input
              type="url"
              value={formData.image_url}
              onChange={(e) => setFormData({ ...formData, image_url: e.target.value })}
              className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500"
            />
          </div>

          {product && (
            <div className="flex items-center">
              <input
                type="checkbox"
                id="is_available"
                checked={formData.is_available}
                onChange={(e) => setFormData({ ...formData, is_available: e.target.checked })}
                className="mr-2"
              />
              <label htmlFor="is_available" className="text-sm font-medium text-gray-700">
                販売中
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
              {submitting ? '処理中...' : product ? '更新' : '作成'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
