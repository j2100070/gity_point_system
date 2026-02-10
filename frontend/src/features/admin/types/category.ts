// カテゴリ関連の型定義

export interface Category {
    ID: string;
    Name: string;
    Code: string;
    Description: string;
    DisplayOrder: number;
    IsActive: boolean;
    CreatedAt: string;
    UpdatedAt: string;
}

export interface CreateCategoryRequest {
    name: string;
    code: string;
    description?: string;
    display_order?: number;
}

export interface UpdateCategoryRequest {
    name: string;
    description?: string;
    display_order?: number;
    is_active: boolean;
}
