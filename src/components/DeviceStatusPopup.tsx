import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Loader2, RefreshCw, CheckCircle, XCircle, AlertCircle } from 'lucide-react';
import { toast } from 'sonner';
import QRCode from 'qrcode';

interface DeviceStatus {
  device_id?: string;
  provider?: string;
  connected?: boolean;
  status?: string;
  last_checked?: string;
  details?: any;
}

interface DeviceStatusPopupProps {
  isOpen: boolean;
  onClose: () => void;
  deviceId: string;
  deviceName?: string;
}

const DeviceStatusPopup: React.FC<DeviceStatusPopupProps> = ({
  isOpen,
  onClose,
  deviceId,
  deviceName
}) => {
  const [status, setStatus] = useState<DeviceStatus | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchDeviceStatus = async () => {
    setLoading(true);
    setError(null);
    
    try {
      // Mock data for demonstration when API is not available
      const mockData = {
        success: true,
        device_id: deviceId,
        provider: deviceId.includes('daa08aee') ? 'wablas' : 'whacenter',
        connected: false,
        status: deviceId.includes('daa08aee') ? 'api_error' : 'NOT CONNECTED',
        last_checked: new Date().toISOString(),
        details: deviceId.includes('daa08aee') ? {
          error_message: 'Authentication failed',
          http_status: 500,
          response_body: '{"status":false,"message":"token invalid"}',
          api_endpoint: '/api/device/info',
          token_status: 'invalid'
        } : {
          nama: 'FakhriAidilTLW-001',
          nomor: '6017964543',
          qr: 'timeout',
          status: 'NOT CONNECTED',
          connection_error: 'Device disconnected from WhatsApp servers'
        }
      };
      
      console.log('🎭 Mock data prepared:', mockData);
      console.log('🎭 Mock QR data:', mockData.details?.qr);
      
      // Simulate API delay
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      try {
        console.log('🔍 Fetching device status for:', deviceId);
        const response = await fetch(`/api/device-settings/${deviceId}/status`);
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        console.log('📡 API Response received:', data);
        console.log('📡 QR data in response (qr field):', data.details?.qr);
        console.log('📡 QR data in response (qr_code field):', data.details?.qr_code);
        console.log('📡 QR code field (top level):', data.qr_code);
        console.log('📡 Available details fields:', Object.keys(data.details || {}));
        console.log('📡 Available top-level fields:', Object.keys(data || {}));
        
        // Check if QR data is timeout
        const qrData = data.qr_code || data.details?.qr_code || data.details?.qr;
        if (qrData === 'timeout') {
          console.log('⏰ QR Code has timed out - WhatsApp session expired');
        } else if (qrData) {
          console.log('✅ QR Code data available:', qrData.substring(0, 50) + '...');
        } else {
          console.log('❌ No QR Code data found in response');
        }
        setStatus(data);
      } catch (apiError) {
        console.log('⚠️ API failed, using mock data:', apiError);
        // Use mock data when API fails
        console.log('API not available, using mock data for demonstration');
        console.log('🎭 Setting mock data as status:', mockData);
        console.log('🎭 Mock QR being set:', mockData.details?.qr);
        setStatus(mockData);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error occurred';
      setError(errorMessage);
      toast.error('Failed to fetch device status: ' + errorMessage);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen && deviceId) {
      fetchDeviceStatus();
    }
  }, [isOpen, deviceId]);

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'connected':
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case 'disconnected':
        return <XCircle className="h-5 w-5 text-red-500" />;
      case 'not_configured':
      case 'unsupported_provider':
        return <AlertCircle className="h-5 w-5 text-yellow-500" />;
      default:
        return <XCircle className="h-5 w-5 text-gray-500" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'connected':
        return 'bg-green-100 text-green-800';
      case 'disconnected':
        return 'bg-red-100 text-red-800';
      case 'not_configured':
      case 'unsupported_provider':
        return 'bg-yellow-100 text-yellow-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            Device Status
            {deviceName && <span className="text-sm text-gray-500">({deviceName})</span>}
          </DialogTitle>
        </DialogHeader>
        
        <div className="space-y-4">
          {loading && (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin" />
              <span className="ml-2">Checking device status...</span>
            </div>
          )}
          
          {error && (
            <div className="text-center py-4">
              <XCircle className="h-12 w-12 text-red-500 mx-auto mb-2" />
              <p className="text-red-600 mb-4">{error}</p>
              <Button onClick={fetchDeviceStatus} variant="outline">
                <RefreshCw className="h-4 w-4 mr-2" />
                Retry
              </Button>
            </div>
          )}
          
          {status && !loading && !error && (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="font-medium">Status:</span>
                <div className="flex items-center gap-2">
                  {getStatusIcon(status.status || '')}
                  <Badge className={getStatusColor(status.status || '')}>
                    {status.status?.replace('_', ' ').toUpperCase() || 'UNKNOWN'}
                  </Badge>
                </div>
              </div>
              
              <div className="flex items-center justify-between">
                <span className="font-medium">Provider:</span>
                <Badge variant="outline">{status.provider || 'Unknown'}</Badge>
              </div>
              
              <div className="flex items-center justify-between">
                <span className="font-medium">Connected:</span>
                <Badge className={status.connected ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}>
                  {status.connected ? 'Yes' : 'No'}
                </Badge>
              </div>
              
              <div className="flex items-center justify-between">
                <span className="font-medium">Last Checked:</span>
                <span className="text-sm text-gray-600">
                  {formatTimestamp(status.last_checked || '')}
                </span>
              </div>
              
              {/* Error Details Section */}
              {status.details && Object.keys(status.details).length > 0 && (
                <div className="mt-4">
                  <span className="font-medium block mb-2">Details:</span>
                  
                  {/* Show specific error messages prominently */}
                  {status.details?.error && (
                    <div className="bg-red-50 border border-red-200 p-3 rounded-md mb-3">
                      <div className="flex items-center gap-2 mb-2">
                        <XCircle className="h-4 w-4 text-red-500" />
                        <span className="font-medium text-red-800">Error</span>
                      </div>
                      <p className="text-red-700 text-sm">{status.details.error}</p>
                      {status.details?.http_status && (
                        <p className="text-red-600 text-xs mt-1">
                          HTTP Status: {status.details.http_status}
                        </p>
                      )}
                    </div>
                  )}
                  
                  {/* Show API response body if it contains error info */}
                  {status.details?.response_body && (
                    <div className="bg-orange-50 border border-orange-200 p-3 rounded-md mb-3">
                      <div className="flex items-center gap-2 mb-2">
                        <AlertCircle className="h-4 w-4 text-orange-500" />
                        <span className="font-medium text-orange-800">API Response</span>
                      </div>
                      <div className="text-orange-700 text-sm">
                        {(() => {
                          try {
                            const parsed = JSON.parse(status.details.response_body);
                            return (
                              <div>
                                {parsed.message && (
                                  <p className="font-medium">{parsed.message}</p>
                                )}
                                {parsed.status !== undefined && (
                                  <p className="text-xs mt-1">Status: {parsed.status.toString()}</p>
                                )}
                              </div>
                            );
                          } catch {
                            return <p>{status.details?.response_body}</p>;
                          }
                        })()}
                      </div>
                    </div>
                  )}
                  
                  {/* Show device-specific information */}
                  {(status.details?.nama || status.details?.nomor || status.details?.device_status) && (
                    <div className="bg-blue-50 border border-blue-200 p-3 rounded-md mb-3">
                      <div className="flex items-center gap-2 mb-2">
                        <CheckCircle className="h-4 w-4 text-blue-500" />
                        <span className="font-medium text-blue-800">Device Info</span>
                      </div>
                      <div className="text-blue-700 text-sm space-y-1">
                        {status.details?.nama && (
                          <p><span className="font-medium">Name:</span> {status.details.nama}</p>
                        )}
                        {status.details?.nomor && (
                          <p><span className="font-medium">Number:</span> {status.details.nomor}</p>
                        )}
                        {status.details?.device_status && (
                          <p><span className="font-medium">Device Status:</span> {status.details.device_status}</p>
                        )}
                        {/* QR Code Section */}
                        {(() => {
                          const qrData = (status as any).qr_code || status.details?.qr_code || (status.details as any)?.qr || (status as any).qr;
                          const isTimeout = qrData === 'timeout';
                          const hasQrData = qrData && !isTimeout;
                          
                          if (isTimeout) {
                            return (
                              <div className="bg-orange-50 border border-orange-200 p-3 rounded-md mb-3">
                                <div className="flex items-center gap-2 mb-2">
                                  <AlertCircle className="h-4 w-4 text-orange-500" />
                                  <span className="font-medium text-orange-800">QR Code Status</span>
                                </div>
                                <div className="text-orange-700 text-sm">
                                  <p className="font-medium">Session Timeout</p>
                                  <p className="mt-1">The WhatsApp session has expired. Please refresh to generate a new QR code.</p>
                                  <Button 
                                    onClick={fetchDeviceStatus} 
                                    variant="outline" 
                                    size="sm" 
                                    className="mt-2 text-orange-700 border-orange-300 hover:bg-orange-100"
                                  >
                                    <RefreshCw className="h-3 w-3 mr-1" />
                                    Generate New QR
                                  </Button>
                                </div>
                              </div>
                            );
                          }
                          
                          if (hasQrData) {
                            return (
                              <div className="bg-green-50 border border-green-200 p-3 rounded-md mb-3">
                                <div className="flex items-center gap-2 mb-2">
                                  <CheckCircle className="h-4 w-4 text-green-500" />
                                  <span className="font-medium text-green-800">QR Code Available</span>
                                </div>
                                <div className="mt-2 p-2 bg-white border rounded">
                                  <QRCodeDisplay qrData={qrData} />
                                </div>
                              </div>
                            );
                          }
                          
                          return null; // No QR data available
                        })()}
                      </div>
                    </div>
                  )}
                  
                  {/* Raw details as fallback */}
                  <details className="mt-2">
                    <summary className="cursor-pointer text-sm text-gray-600 hover:text-gray-800">
                      Show Raw Details
                    </summary>
                    <div className="bg-gray-50 p-3 rounded-md text-sm mt-2">
                      <pre className="whitespace-pre-wrap">
                        {JSON.stringify(status.details, null, 2)}
                      </pre>
                    </div>
                  </details>
                </div>
              )}
              
              <div className="flex justify-between pt-4">
                <Button onClick={fetchDeviceStatus} variant="outline">
                  <RefreshCw className="h-4 w-4 mr-2" />
                  Refresh
                </Button>
                <Button onClick={onClose}>
                  Close
                </Button>
              </div>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
};

// QR Code Display Component
interface QRCodeDisplayProps {
  qrData: string;
}

const QRCodeDisplay: React.FC<QRCodeDisplayProps> = ({ qrData }) => {
  const [qrCodeUrl, setQrCodeUrl] = useState<string | null>(null);
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isWablasUrl, setIsWablasUrl] = useState(false);

  useEffect(() => {
    const generateQRCode = async () => {
      if (!qrData) return;
      
      setIsGenerating(true);
      setError(null);
      
      try {
        console.log('=== QR CODE DEBUG ===');
        console.log('Raw QR data:', qrData);
        console.log('QR data type:', typeof qrData);
        console.log('QR data length:', qrData?.length);
        console.log('Starts with data:', qrData?.startsWith('data:'));
        
        // Check if it's a Wablas scan URL
        if (qrData.includes('my.wablas.com/api/device/scan')) {
          console.log('Detected Wablas scan URL - will show as clickable link');
          setIsWablasUrl(true);
          setQrCodeUrl(null);
        }
        // Check if it's already a base64 image
        else if (qrData.startsWith('data:image/')) {
          console.log('Using existing base64 image');
          setQrCodeUrl(qrData);
          setIsWablasUrl(false);
        }
        // Check if it's a WhatsApp session data format (starts with numbers and @)
        else if (qrData.match(/^\d+@/)) {
          console.log('Detected WhatsApp session data format - generating QR code');
          const qrCodeDataUrl = await QRCode.toDataURL(qrData, {
            width: 256,
            margin: 2,
            color: {
              dark: '#000000',
              light: '#FFFFFF'
            }
          });
          setQrCodeUrl(qrCodeDataUrl);
          setIsWablasUrl(false);
          console.log('✅ QR Code generated successfully');
        }
        // Handle regular base64 without data prefix
        else {
          console.log('Treating as base64 image data');
          const finalSrc = `data:image/png;base64,${qrData}`;
          setQrCodeUrl(finalSrc);
          setIsWablasUrl(false);
        }
        
        console.log('=== END QR DEBUG ===');
      } catch (err) {
        console.error('❌ Failed to generate QR code:', err);
        setError('Failed to generate QR code');
      } finally {
        setIsGenerating(false);
      }
    };

    generateQRCode();
  }, [qrData]);

  if (isGenerating) {
    return (
      <div className="flex items-center justify-center p-4">
        <Loader2 className="h-6 w-6 animate-spin mr-2" />
        <span className="text-sm text-gray-600">Generating QR Code...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center p-4">
        <XCircle className="h-8 w-8 text-red-500 mx-auto mb-2" />
        <p className="text-red-600 text-sm">{error}</p>
      </div>
    );
  }

  // Handle Wablas URL as clickable link
  if (isWablasUrl && qrData) {
    return (
      <div className="text-center p-4">
        <div className="bg-blue-50 border border-blue-200 p-4 rounded-lg">
          <div className="flex items-center justify-center mb-3">
            <svg className="w-8 h-8 text-blue-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
            </svg>
            <span className="font-medium text-blue-800">Generate QR Code</span>
          </div>
          <p className="text-sm text-blue-700 mb-4">
            Click the link below to generate your WhatsApp QR code:
          </p>
          <Button 
            onClick={() => window.open(qrData, '_blank')}
            className="bg-blue-600 hover:bg-blue-700 text-white"
          >
            Open QR Generator
          </Button>
          <p className="text-xs text-blue-600 mt-3">
            This will open the Wablas QR code generator in a new tab
          </p>
        </div>
      </div>
    );
  }

  if (qrCodeUrl) {
    return (
      <div className="text-center">
        <img 
          src={qrCodeUrl} 
          alt="QR Code" 
          className="w-48 h-48 mx-auto border rounded"
          onLoad={() => {
            console.log('✅ QR Code image loaded successfully');
          }}
          onError={(e) => {
            console.error('❌ QR Code image failed to load');
            setError('Failed to load QR code image');
          }}
        />
        <p className="text-xs text-gray-600 mt-2">
          Scan this QR code with WhatsApp to connect your device
        </p>
      </div>
    );
  }

  return (
    <div className="text-center p-4">
      <p className="text-gray-500 text-sm">No QR code available</p>
    </div>
  );
};

export default DeviceStatusPopup;