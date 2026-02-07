import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getProducts, exchangeProduct } from '../api/productsApi';
import type { Product } from '../types';
import { useAuthStore } from '@/shared/stores/authStore';

export const ExchangePage: React.FC = () => {
  const { productId } = useParams<{ productId: string }>();
  const navigate = useNavigate();
  const { user } = useAuthStore();

  const [product, setProduct] = useState<Product | null>(null);
  const [quantity, setQuantity] = useState(1);
  const [notes, setNotes] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadProduct();
  }, [productId]);

  const loadProduct = async () => {
    if (!productId) return;

    try {
      setLoading(true);
      const data = await getProducts();
      const foundProduct = data.Products.find((p) => p.ID === productId);
      if (foundProduct) {
        setProduct(foundProduct);
      } else {
        setError('商品が見つかりませんでした');
      }
    } catch (err: any) {
      setError(err.response?.data?.error || '商品の読み込みに失敗しました');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!productId || !product) return;

    // バリデーション
    if (quantity <= 0) {
      setError('数量は1以上を指定してください');
      return;
    }

    if (product.Stock !== -1 && quantity > product.Stock) {
      setError('在庫が不足しています');
      return;
    }

    const totalPoints = product.Price * quantity;
    if (user && user.balance < totalPoints) {
      setError(`ポイントが不足しています（必要: ${totalPoints}pt、現在: ${user.balance}pt）`);
      return;
    }

    try {
      setSubmitting(true);
      setError(null);
      await exchangeProduct({
        product_id: productId,
        quantity,
        notes: notes || undefined,
      });

      alert('商品交換が完了しました！');
      navigate('/products/exchanges');
    } catch (err: any) {
      setError(err.response?.data?.error || '交換に失敗しました');
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <div className="text-lg">読み込み中...</div>
      </div>
    );
  }

  if (!product) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-8">
        <div className="bg-red-50 border border-red-200 text-red-600 rounded-md p-4">
          {error || '商品が見つかりませんでした'}
        </div>
        <button
          onClick={() => navigate('/products')}
          className="mt-4 text-blue-600 hover:underline"
        >
          ← 商品一覧に戻る
        </button>
      </div>
    );
  }

  const totalPoints = product.Price * quantity;
  const canExchange =
    product.IsAvailable &&
    (product.Stock === -1 || product.Stock >= quantity) &&
    user &&
    user.balance >= totalPoints;

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <button
        onClick={() => navigate('/products')}
        className="mb-6 text-blue-600 hover:underline flex items-center"
      >
        ← 商品一覧に戻る
      </button>

      <div className="bg-white rounded-lg shadow-md overflow-hidden">
        <div className="md:flex">
          {/* 商品画像 */}
          <div className="md:w-1/2">
            <div className="h-96 bg-gray-200 flex items-center justify-center">
              {product.ImageURL ? (
                <img
                  src={product.ImageURL}
                  alt={product.Name}
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="text-gray-400 text-6xl">📦</div>
              )}
            </div>
          </div>

          {/* 商品情報と交換フォーム */}
          <div className="md:w-1/2 p-8">
            <h1 className="text-3xl font-bold text-gray-900 mb-4">{product.Name}</h1>
            <p className="text-gray-600 mb-6">{product.Description}</p>

            <div className="mb-6 space-y-3">
              <div className="flex justify-between">
                <span className="text-gray-600">価格</span>
                <span className="text-2xl font-bold text-blue-600">
                  {product.Price} pt
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-600">在庫</span>
                <span className="font-semibold">
                  {product.Stock === -1 ? '無制限' : `${product.Stock}個`}
                </span>
              </div>
              {user && (
                <div className="flex justify-between">
                  <span className="text-gray-600">所持ポイント</span>
                  <span className="font-semibold">{user.balance} pt</span>
                </div>
              )}
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
              {/* 数量 */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  数量
                </label>
                <input
                  type="number"
                  min="1"
                  max={product.Stock === -1 ? undefined : product.Stock}
                  value={quantity}
                  onChange={(e) => setQuantity(parseInt(e.target.value) || 1)}
                  className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              {/* 備考 */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  備考（任意）
                </label>
                <textarea
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  placeholder="受取場所や希望日時など"
                  rows={3}
                  className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              {/* 合計 */}
              <div className="bg-gray-50 p-4 rounded-md">
                <div className="flex justify-between text-lg font-semibold">
                  <span>合計</span>
                  <span className="text-blue-600">{totalPoints} pt</span>
                </div>
              </div>

              {error && (
                <div className="p-4 bg-red-50 border border-red-200 text-red-600 rounded-md text-sm">
                  {error}
                </div>
              )}

              {/* 交換ボタン */}
              <button
                type="submit"
                disabled={!canExchange || submitting}
                className={`w-full py-3 px-4 rounded-md font-semibold text-white transition-colors ${
                  canExchange && !submitting
                    ? 'bg-blue-600 hover:bg-blue-700'
                    : 'bg-gray-300 cursor-not-allowed'
                }`}
              >
                {submitting ? '交換中...' : 'ポイントと交換する'}
              </button>

              {!canExchange && !submitting && (
                <p className="text-sm text-red-600 text-center">
                  {!product.IsAvailable
                    ? '現在交換できません'
                    : product.Stock !== -1 && product.Stock < quantity
                    ? '在庫が不足しています'
                    : user && user.balance < totalPoints
                    ? 'ポイントが不足しています'
                    : ''}
                </p>
              )}
            </form>
          </div>
        </div>
      </div>
    </div>
  );
};
