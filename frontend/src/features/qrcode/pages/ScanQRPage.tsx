import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { QRCodeRepository } from '@/infrastructure/api/repositories/QRCodeRepository';

const qrRepository = new QRCodeRepository();

export const ScanQRPage: React.FC = () => {
  const [qrData, setQrData] = useState('');
  const [amount, setAmount] = useState('');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const generateIdempotencyKey = () => {
    return `qr-scan-${Date.now()}-${Math.random().toString(36).substring(7)}`;
  };

  const handleScan = async () => {
    if (!qrData.trim()) {
      setError('QRコードデータを入力してください');
      return;
    }

    setError('');
    setLoading(true);
    try {
      await qrRepository.scanQR({
        qr_code: qrData.trim(),
        amount: amount ? parseInt(amount) : undefined,
        idempotency_key: generateIdempotencyKey(),
      });
      setSuccess(true);
      setTimeout(() => {
        navigate('/dashboard');
      }, 2000);
    } catch (err: any) {
      setError(err.response?.data?.error || 'QRコードスキャンに失敗しました');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-md mx-auto space-y-6 pb-20 md:pb-6">
      <div className="flex items-center mb-6">
        <button
          onClick={() => navigate(-1)}
          className="mr-4 p-2 hover:bg-gray-100 rounded-full"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <h1 className="text-2xl font-bold">QRコードスキャン</h1>
      </div>

      {success ? (
        <div className="bg-white rounded-xl shadow p-8 text-center space-y-4">
          <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto">
            <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 className="text-xl font-bold text-gray-900">送信完了!</h2>
          <p className="text-gray-600">ダッシュボードに戻ります...</p>
        </div>
      ) : (
        <div className="bg-white rounded-xl shadow p-6 space-y-4">
          <div className="bg-gray-100 rounded-lg p-8 text-center">
            <svg className="w-32 h-32 mx-auto text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
            </svg>
            <p className="mt-4 text-sm text-gray-600">
              カメラでQRコードをスキャン<br />
              または手動入力
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              QRコードデータ
            </label>
            <textarea
              value={qrData}
              onChange={(e) => setQrData(e.target.value)}
              placeholder="receive:CODE:500 または send:CODE:1000"
              rows={3}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            />
            <p className="mt-2 text-xs text-gray-500">
              例: receive:abc123:500 (受取用、500P指定)
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              金額 (金額未指定QRの場合のみ)
            </label>
            <div className="relative">
              <input
                type="number"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="金額が指定されている場合は不要"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              />
              <span className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-500">P</span>
            </div>
          </div>

          {error && (
            <div className="rounded-lg bg-red-50 p-4">
              <div className="text-sm text-red-800">{error}</div>
            </div>
          )}

          <button
            onClick={handleScan}
            disabled={loading}
            className="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? '処理中...' : 'スキャン実行'}
          </button>
        </div>
      )}
    </div>
  );
};
