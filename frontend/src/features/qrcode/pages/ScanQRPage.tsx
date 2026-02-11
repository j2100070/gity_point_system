import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Scanner } from '@yudiel/react-qr-scanner';
import { TransferRequestRepository } from '@/infrastructure/api/repositories/TransferRequestRepository';

const transferRequestRepository = new TransferRequestRepository();

export const ScanQRPage: React.FC = () => {
  const [qrData, setQrData] = useState('');
  const [amount, setAmount] = useState('');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState('');
  const [isCameraActive, setIsCameraActive] = useState(true);
  const navigate = useNavigate();

  const generateIdempotencyKey = () => {
    return `transfer-request-${Date.now()}-${Math.random().toString(36).substring(7)}`;
  };

  // QRコードのタイプを判定
  const parseQRCode = (code: string) => {
    if (code.startsWith('user:')) {
      // 個人QRコード: user:UUID
      const userId = code.replace('user:', '');
      return { type: 'user', userId };
    }
    return { type: 'unknown' };
  };

  const handleScan = async () => {
    if (!qrData.trim()) {
      setError('QRコードデータを入力してください');
      return;
    }

    if (!amount || parseInt(amount) <= 0) {
      setError('送金額を入力してください');
      return;
    }

    setError('');
    setLoading(true);
    try {
      const parsed = parseQRCode(qrData.trim());

      if (parsed.type === 'user') {
        // 個人QRコード → 送金リクエスト作成
        await transferRequestRepository.createTransferRequest({
          to_user_id: parsed.userId,
          amount: parseInt(amount),
          message: message,
          idempotency_key: generateIdempotencyKey(),
        });
        setSuccess(true);
        setTimeout(() => {
          navigate('/transfer-requests');
        }, 2000);
      } else {
        setError('無効なQRコードです');
      }
    } catch (err: any) {
      setError(err.response?.data?.error || '送金リクエストの作成に失敗しました');
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
          <h2 className="text-xl font-bold text-gray-900">送金リクエストを送信しました!</h2>
          <p className="text-gray-600">相手が承認すると送金が完了します</p>
          <p className="text-sm text-gray-500">リクエスト一覧画面に移動します...</p>
        </div>
      ) : (
        <div className="bg-white rounded-xl shadow p-6 space-y-4">
          {/* カメラによるQRスキャン */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-gray-700">カメラでQRコードを読み取る</span>
              <button
                type="button"
                onClick={() => setIsCameraActive((prev) => !prev)}
                className="text-xs px-3 py-1 rounded-full border border-gray-300 text-gray-600 hover:bg-gray-50"
              >
                {isCameraActive ? 'カメラ停止' : 'カメラ起動'}
              </button>
            </div>
            {isCameraActive && (
              <div className="overflow-hidden rounded-xl border border-gray-200 bg-black">
                <Scanner
                  onScan={(detectedCodes) => {
                    if (detectedCodes && detectedCodes.length > 0) {
                      setQrData(detectedCodes[0].rawValue);
                    }
                  }}
                  onError={(err) => {
                    console.error(err);
                    setError('カメラの起動に失敗しました。ブラウザの権限設定を確認してください。');
                  }}
                  constraints={{ facingMode: 'environment' }}
                  styles={{
                    container: { width: '100%' }
                  }}
                />
              </div>
            )}
            <p className="mt-1 text-xs text-gray-500">
              相手のマイQRコードをスキャンしてください
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              QRコードデータ
            </label>
            <textarea
              value={qrData}
              onChange={(e) => setQrData(e.target.value)}
              placeholder="user:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
              rows={3}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            />
            <p className="mt-2 text-xs text-gray-500">
              例: user:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              送金額 <span className="text-red-500">*</span>
            </label>
            <div className="relative">
              <input
                type="number"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="送りたいポイント数を入力"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                required
              />
              <span className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-500">P</span>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              メッセージ（任意）
            </label>
            <textarea
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder="メモを追加できます"
              rows={2}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            />
          </div>

          {error && (
            <div className="rounded-lg bg-red-50 p-4">
              <div className="text-sm text-red-800">{error}</div>
            </div>
          )}

          {/* 説明 */}
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="flex items-start">
              <svg className="w-5 h-5 text-blue-600 mr-2 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <div className="text-xs text-blue-800">
                <div className="font-medium mb-1">送金の流れ</div>
                <ol className="list-decimal list-inside space-y-1">
                  <li>相手のマイQRコードをスキャン</li>
                  <li>送りたいポイント数を入力</li>
                  <li>送金リクエストを送信</li>
                  <li>相手が「受け取る」を押すと完了</li>
                </ol>
              </div>
            </div>
          </div>

          <button
            onClick={handleScan}
            disabled={loading}
            className="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? '送信中...' : '送金リクエストを送信'}
          </button>
        </div>
      )}
    </div>
  );
};
