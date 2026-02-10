import { axiosInstance } from '@/infrastructure/api/client';
import type { Category, CreateCategoryRequest, UpdateCategoryRequest } from '../types/category';

const API_BASE_URL = '/api';

// カテゴリ一覧取得（公開API）
export const getCategories = async (activeOnly: boolean = true): Promise<{ categories: Category[]; total: number }> => {
    const response = await axiosInstance.get(`${API_BASE_URL}/categories`, {
        params: { active_only: activeOnly }
    });
    return response.data;
};

// === 管理者用API ===

// カテゴリ作成
export const createCategory = async (request: CreateCategoryRequest): Promise<{ category: Category }> => {
    const response = await axiosInstance.post(`${API_BASE_URL}/admin/categories`, request);
    return response.data;
};

// カテゴリ更新
export const updateCategory = async (
    categoryId: string,
    request: UpdateCategoryRequest
): Promise<{ category: Category }> => {
    const response = await axiosInstance.put(`${API_BASE_URL}/admin/categories/${categoryId}`, request);
    return response.data;
};

// カテゴリ削除
export const deleteCategory = async (categoryId: string): Promise<void> => {
    await axiosInstance.delete(`${API_BASE_URL}/admin/categories/${categoryId}`);
};
