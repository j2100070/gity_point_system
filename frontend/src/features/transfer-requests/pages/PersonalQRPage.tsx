import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { QRCodeSVG } from 'qrcode.react';
import { TransferRequestRepository } from '@/infrastructure/api/repositories/TransferRequestRepository';

const transferRequestRepository = new TransferRequestRepository();

export const PersonalQRPage: React.FC = () => {
  const [qrCode, setQrCode] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [username, setUsername] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    loadPersonalQR();
  }, []);

  const loadPersonalQR = async () => {
    try {
      const response = await transferRequestRepository.getPersonalQRCode();
      setQrCode(response.qr_code);
      setDisplayName(response.display_name);
      setUsername(response.username);
    } catch (err: any) {
      setError(err.response?.data?.error || '個人QRコードの取得に失敗しました');
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
        <h1 className="text-2xl font-bold">マイQRコード</h1>
      </div>

      {loading ? (
        <div className="flex justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
        </div>
      ) : error ? (
        <div className="bg-white rounded-xl shadow p-6">
          <div className="rounded-lg bg-red-50 p-4">
            <div className="text-sm text-red-800">{error}</div>
          </div>
        </div>
      ) : (
        <div className="bg-white rounded-xl shadow p-6 space-y-6">
          {/* QRコード表示 */}
          <div className="text-center">
            <div className="inline-block p-6 bg-white border-4 border-primary-500 rounded-2xl">
              <QRCodeSVG value={qrCode} size={280} level="H" />
            </div>
          </div>

          {/* ユーザー情報 */}
          <div className="text-center space-y-2">
            <div className="text-2xl font-bold text-gray-900">{displayName}</div>
            <div className="text-sm text-gray-500">@{username}</div>
          </div>

          {/* 説明 */}
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="flex items-start">
              <svg className="w-5 h-5 text-blue-600 mr-2 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <div className="text-sm text-blue-800">
                <div className="font-medium mb-1">このQRコードの使い方</div>
                <ul className="list-disc list-inside space-y-1 text-xs">
                  <li>このQRコードは常に有効です（有効期限なし）</li>
                  <li>送金したい相手にこのQRコードを見せてください</li>
                  <li>相手がQRコードをスキャンして金額を入力します</li>
                  <li>あなたが「受け取る」ボタンを押すと送金が完了します</li>
                </ul>
              </div>
            </div>
          </div>

          {/* QRコードテキスト（デバッグ用・コピー可能） */}
          <div className="bg-gray-50 rounded-lg p-4">
            <div className="text-xs text-gray-500 mb-2">QRコードデータ</div>
            <div className="text-xs font-mono bg-white p-2 rounded border border-gray-200 break-all">
              {qrCode}
            </div>
          </div>

          {/* アクションボタン */}
          <div className="space-y-3">
            <button
              onClick={() => navigate('/transfer-requests')}
              className="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700"
            >
              送金リクエストを確認
            </button>
            <button
              onClick={() => navigate('/dashboard')}
              className="w-full border border-gray-300 text-gray-700 py-3 rounded-lg font-medium hover:bg-gray-50"
            >
              ダッシュボードに戻る
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
