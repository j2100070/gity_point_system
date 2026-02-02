import axios, { AxiosInstance, InternalAxiosRequestConfig } from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

console.log('API Base URL:', API_BASE_URL);

class ApiClient {
  private client: AxiosInstance;
  private csrfToken: string = '';

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      withCredentials: true, // Session cookie送信のため必須
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // リクエストインターセプター: CSRF Tokenを自動付与
    this.client.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        // POST, PUT, DELETE, PATCHにCSRFトークンを付与
        if (['post', 'put', 'delete', 'patch'].includes(config.method?.toLowerCase() || '')) {
          if (this.csrfToken) {
            config.headers['X-CSRF-Token'] = this.csrfToken;
          }
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // レスポンスインターセプター: CSRF Tokenの保存
    this.client.interceptors.response.use(
      (response) => {
        // ログイン成功時などにCSRFトークンを保存
        if (response.data?.csrf_token) {
          this.setCsrfToken(response.data.csrf_token);
        }
        return response;
      },
      (error) => {
        // 401エラー時はログイン画面へリダイレクト（後でルーターで実装）
        if (error.response?.status === 401) {
          this.csrfToken = '';
          // ログインページへのリダイレクトはコンポーネント側で処理
        }
        return Promise.reject(error);
      }
    );
  }

  public setCsrfToken(token: string) {
    this.csrfToken = token;
    // LocalStorageにも保存（リロード時のため）
    if (typeof window !== 'undefined') {
      localStorage.setItem('csrf_token', token);
    }
  }

  public getCsrfToken(): string {
    if (!this.csrfToken && typeof window !== 'undefined') {
      this.csrfToken = localStorage.getItem('csrf_token') || '';
    }
    return this.csrfToken;
  }

  public clearCsrfToken() {
    this.csrfToken = '';
    if (typeof window !== 'undefined') {
      localStorage.removeItem('csrf_token');
    }
  }

  public getClient(): AxiosInstance {
    return this.client;
  }
}

// シングルトンインスタンス
export const apiClient = new ApiClient();
export const axiosInstance = apiClient.getClient();
