import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { QRCodeSVG } from 'qrcode.react';
import { QRCodeRepository } from '@/infrastructure/api/repositories/QRCodeRepository';
import { QRCode } from '@/core/domain/QRCode';

const qrRepository = new QRCodeRepository();

export const ReceiveQRPage: React.FC = () => {
  const [amount, setAmount] = useState('');
  const [qrCode, setQrCode] = useState<QRCode | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleGenerate = async () => {
    setError('');
    setLoading(true);
    try {
      const response = await qrRepository.generateReceiveQR({
        amount: amount ? parseInt(amount) : undefined,
      });
      setQrCode(response.qr_code);
    } catch (err: any) {
      setError(err.response?.data?.error || 'QRコード生成に失敗しました');
    } finally {
      setLoading(false);
    }
  };

  const formatExpiryTime = (expiresAt: string) => {
    const date = new Date(expiresAt);
    return date.toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' });
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
        <h1 className="text-2xl font-bold">受取用QRコード</h1>
      </div>

      {!qrCode ? (
        <div className="bg-white rounded-xl shadow p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              金額 (オプション)
            </label>
            <div className="relative">
              <input
                type="number"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="金額未指定の場合は空欄"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              />
              <span className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-500">P</span>
            </div>
            <p className="mt-2 text-xs text-gray-500">
              金額を指定しない場合、送信者が金額を入力します
            </p>
          </div>

          {error && (
            <div className="rounded-lg bg-red-50 p-4">
              <div className="text-sm text-red-800">{error}</div>
            </div>
          )}

          <button
            onClick={handleGenerate}
            disabled={loading}
            className="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? 'QRコード生成中...' : 'QRコード生成'}
          </button>
        </div>
      ) : (
        <div className="bg-white rounded-xl shadow p-6 space-y-6">
          <div className="text-center">
            <div className="inline-block p-4 bg-white border-4 border-primary-500 rounded-2xl">
              <QRCodeSVG value={qrCode.qr_data} size={256} level="H" />
            </div>
          </div>

          <div className="space-y-3">
            {qrCode.amount && (
              <div className="text-center">
                <div className="text-sm text-gray-600">金額</div>
                <div className="text-3xl font-bold text-primary-600">
                  {qrCode.amount.toLocaleString()} P
                </div>
              </div>
            )}

            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
              <div className="flex items-start">
                <svg className="w-5 h-5 text-yellow-600 mr-2 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <div className="text-sm text-yellow-800">
                  <div className="font-medium mb-1">有効期限</div>
                  <div>{formatExpiryTime(qrCode.expires_at)} まで (5分間)</div>
                </div>
              </div>
            </div>

            <div className="text-center text-sm text-gray-600">
              このQRコードを相手にスキャンしてもらってください
            </div>
          </div>

          <button
            onClick={() => setQrCode(null)}
            className="w-full border border-gray-300 text-gray-700 py-3 rounded-lg font-medium hover:bg-gray-50"
          >
            新しいQRコードを生成
          </button>
        </div>
      )}
    </div>
  );
};
