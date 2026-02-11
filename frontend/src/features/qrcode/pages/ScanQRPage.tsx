import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Scanner } from '@yudiel/react-qr-scanner';

export const ScanQRPage: React.FC = () => {
  const [error, setError] = useState('');
  const [isCameraActive, setIsCameraActive] = useState(true);
  const navigate = useNavigate();

  // QRコードのタイプを判定して遷移
  const parseQRCodeAndNavigate = (code: string) => {
    if (code.startsWith('user:')) {
      // 個人QRコード: user:UUID
      const userId = code.replace('user:', '');
      // 確認ページに遷移
      navigate(`/qr/confirm?userId=${userId}`);
      return true;
    }
    return false;
  };

  const handleScanSuccess = (detectedCodes: any[]) => {
    if (detectedCodes && detectedCodes.length > 0) {
      const code = detectedCodes[0].rawValue;
      const success = parseQRCodeAndNavigate(code);
      if (!success) {
        setError('無効なQRコードです。個人QRコードをスキャンしてください。');
      }
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
                onScan={handleScanSuccess}
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
                <li>ユーザー情報を確認</li>
                <li>送りたいポイント数を入力</li>
                <li>送金リクエストを送信</li>
                <li>相手が「受け取る」を押すと完了</li>
              </ol>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
