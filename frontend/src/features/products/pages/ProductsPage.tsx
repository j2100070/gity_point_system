import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { getProducts } from '../api/productsApi';
import type { Product, ProductCategory } from '../types';

const CATEGORIES: { value: ProductCategory | ''; label: string }[] = [
  { value: '', label: 'すべて' },
  { value: 'drink', label: '飲み物' },
  { value: 'snack', label: 'お菓子' },
  { value: 'toy', label: 'おもちゃ' },
  { value: 'other', label: 'その他' },
];

export const ProductsPage: React.FC = () => {
  const navigate = useNavigate();
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedCategory, setSelectedCategory] = useState<ProductCategory | ''>('');

  useEffect(() => {
    loadProducts();
  }, [selectedCategory]);

  const loadProducts = async () => {
    try {
      setLoading(true);
      setError(null);
      const params = {
        category: selectedCategory || undefined,
        available_only: true,
      };
      const data = await getProducts(params);
      setProducts(data.Products || []);
    } catch (err: any) {
      setError(err.response?.data?.error || '商品の読み込みに失敗しました');
    } finally {
      setLoading(false);
    }
  };

  const handleExchange = (productId: string) => {
    navigate(`/products/${productId}/exchange`);
  };

  const getCategoryLabel = (category: ProductCategory) => {
    return CATEGORIES.find((c) => c.value === category)?.label || category;
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <div className="text-lg">読み込み中...</div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-4">商品一覧</h1>
        <p className="text-gray-600">ポイントと交換できる商品一覧です</p>
      </div>

      {/* カテゴリフィルタ */}
      <div className="mb-6">
        <label className="block text-sm font-medium text-gray-700 mb-2">カテゴリ</label>
        <div className="flex flex-wrap gap-2">
          {CATEGORIES.map((cat) => (
            <button
              key={cat.value}
              onClick={() => setSelectedCategory(cat.value as ProductCategory | '')}
              className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${selectedCategory === cat.value
                  ? 'bg-blue-600 text-white'
                  : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
                }`}
            >
              {cat.label}
            </button>
          ))}
        </div>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 text-red-600 rounded-md">
          {error}
        </div>
      )}

      {/* 商品グリッド */}
      {products.length === 0 ? (
        <div className="text-center py-12 text-gray-500">
          商品が見つかりませんでした
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {products.map((product) => (
            <div
              key={product.ID}
              className="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow"
            >
              {/* 商品画像 */}
              <div className="h-48 bg-gray-200 flex items-center justify-center">
                {product.ImageURL ? (
                  <img
                    src={product.ImageURL}
                    alt={product.Name}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="text-gray-400 text-4xl">📦</div>
                )}
              </div>

              {/* 商品情報 */}
              <div className="p-4">
                <div className="mb-2">
                  <span className="inline-block px-2 py-1 text-xs font-semibold text-blue-600 bg-blue-100 rounded">
                    {getCategoryLabel(product.CategoryCode)}
                  </span>
                </div>

                <h3 className="text-lg font-semibold text-gray-900 mb-2">
                  {product.Name}
                </h3>

                <p className="text-sm text-gray-600 mb-4 line-clamp-2">
                  {product.Description}
                </p>

                <div className="flex justify-between items-center mb-4">
                  <div>
                    <span className="text-2xl font-bold text-blue-600">
                      {product.Price}
                    </span>
                    <span className="text-sm text-gray-600 ml-1">pt</span>
                  </div>
                  <div className="text-sm text-gray-600">
                    在庫: {product.Stock === -1 ? '無制限' : product.Stock}
                  </div>
                </div>

                <button
                  onClick={() => handleExchange(product.ID)}
                  disabled={!product.IsAvailable || (product.Stock !== -1 && product.Stock === 0)}
                  className={`w-full py-2 px-4 rounded-md font-medium transition-colors ${product.IsAvailable && (product.Stock === -1 || product.Stock > 0)
                      ? 'bg-blue-600 text-white hover:bg-blue-700'
                      : 'bg-gray-300 text-gray-500 cursor-not-allowed'
                    }`}
                >
                  {product.IsAvailable
                    ? product.Stock === 0
                      ? '在庫切れ'
                      : '交換する'
                    : '販売停止中'}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
