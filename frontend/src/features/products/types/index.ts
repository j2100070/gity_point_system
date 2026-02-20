// カテゴリは動的に管理されるためstringに変更
export type ProductCategory = string;

export interface Product {
  ID: string;
  Name: string;
  Description: string;
  CategoryCode: string; // カテゴリコード（例: drink, snack）
  Price: number;
  Stock: number;
  ImageURL: string;
  IsAvailable: boolean;
  CreatedAt: string;
  UpdatedAt: string;
}

export type ExchangeStatus = 'pending' | 'completed' | 'cancelled' | 'delivered';

export interface ProductExchange {
  ID: string;
  UserID: string;
  ProductID: string;
  Quantity: number;
  PointsUsed: number;
  Status: ExchangeStatus;
  TransactionID: string;
  Notes: string;
  CreatedAt: string;
  CompletedAt?: string;
  DeliveredAt?: string;
}

export interface ExchangeProductRequest {
  product_id: string;
  quantity: number;
  notes?: string;
}

export interface CreateProductRequest {
  name: string;
  description: string;
  category: ProductCategory;
  price: number;
  stock: number;
  image_url?: string;
}

export interface UpdateProductRequest extends CreateProductRequest {
  is_available: boolean;
}
