import { axiosInstance } from '@/infrastructure/api/client';
import type {
  Product,
  ProductExchange,
  ExchangeProductRequest,
  CreateProductRequest,
  UpdateProductRequest,
  ProductCategory,
} from '../types';

const API_BASE_URL = '/api';

// 商品一覧取得（公開API）
export const getProducts = async (params?: {
  category?: ProductCategory;
  available_only?: boolean;
  offset?: number;
  limit?: number;
}): Promise<{ Products: Product[]; Total: number }> => {
  const response = await axiosInstance.get(`${API_BASE_URL}/products`, { params });
  return response.data;
};

// 商品交換
export const exchangeProduct = async (
  request: ExchangeProductRequest
): Promise<{
  Exchange: ProductExchange;
  Product: Product;
  User: any;
  Transaction: any;
}> => {
  const response = await axiosInstance.post(`${API_BASE_URL}/products/exchange`, request);
  return response.data;
};

// 交換履歴取得
export const getExchangeHistory = async (params?: {
  offset?: number;
  limit?: number;
}): Promise<{ Exchanges: ProductExchange[]; Total: number }> => {
  const response = await axiosInstance.get(`${API_BASE_URL}/products/exchanges/history`, { params });
  return response.data;
};

// 交換キャンセル
export const cancelExchange = async (exchangeId: string): Promise<void> => {
  await axiosInstance.post(`${API_BASE_URL}/products/exchanges/${exchangeId}/cancel`);
};

// === 管理者用API ===

// 商品作成
export const createProduct = async (request: CreateProductRequest): Promise<{ Product: Product }> => {
  const response = await axiosInstance.post(`${API_BASE_URL}/admin/products`, request);
  return response.data;
};

// 商品更新
export const updateProduct = async (
  productId: string,
  request: UpdateProductRequest
): Promise<{ Product: Product }> => {
  const response = await axiosInstance.put(`${API_BASE_URL}/admin/products/${productId}`, request);
  return response.data;
};

// 商品削除
export const deleteProduct = async (productId: string): Promise<void> => {
  await axiosInstance.delete(`${API_BASE_URL}/admin/products/${productId}`);
};

// 全交換履歴取得（管理者）
export const getAllExchanges = async (params?: {
  offset?: number;
  limit?: number;
}): Promise<{ Exchanges: ProductExchange[]; Total: number }> => {
  const response = await axiosInstance.get(`${API_BASE_URL}/admin/exchanges`, { params });
  return response.data;
};

// 配達完了マーク
export const markExchangeDelivered = async (exchangeId: string): Promise<void> => {
  await axiosInstance.post(`${API_BASE_URL}/admin/exchanges/${exchangeId}/deliver`);
};
