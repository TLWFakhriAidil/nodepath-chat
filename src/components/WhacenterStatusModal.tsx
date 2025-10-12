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
  qr_code?: string;
  qr?: string;
  details?: any;
}

interface WhacenterStatusModalProps {
  isOpen: boolean;
  onClose: () => void;
  deviceId: string;
  deviceName?: string;
}

interface QRCodeDisplayProps {
  qrData: string;
}

const QRCodeDisplay: React.FC<QRCodeDisplayProps> = ({ qrData }) => {
  const [qrCodeUrl, setQrCodeUrl] = useState<string | null>(null);
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
        
        // Check if it's already a base64 image
        if (qrData.startsWith('data:image/')) {
          console.log('Using existing base64 image');
          setQrCodeUrl(qrData);
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
          console.log('✅ QR Code generated successfully');
        }
        // Handle regular base64 without data prefix
        else {
          console.log('Treating as base64 image data');
          const finalSrc = `data:image/png;base64,${qrData}`;
          setQrCodeUrl(finalSrc);
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

  if (qrCodeUrl) {
    return (
      <div className="text-center">
        <div className="flex justify-center mb-4">
          <div className="bg-white p-4 rounded-lg shadow-sm border">
            <img 
              src={qrCodeUrl} 
              alt="QR Code" 
              className="w-48 h-48 object-contain"
              onLoad={() => {
                console.log('✅ QR Code image loaded successfully');
              }}
              onError={(e) => {
                console.error('❌ QR Code image failed to load');
                setError('Failed to load QR code image');
              }}
            />
          </div>
        </div>
        <p className="text-xs text-gray-600 text-center">
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

const WhacenterStatusModal: React.FC<WhacenterStatusModalProps> = ({
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
      const response = await fetch(`/api/device-settings/${deviceId}/status`);
      const data = await response.json();
      
      if (data.success) {
        setStatus(data.data);
      } else {
        setError('Failed to fetch device status');
      }
    } catch (err) {
      console.error('Error fetching device status:', err);
      setError('Failed to connect to server');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isOpen && deviceId) {
      fetchDeviceStatus();
    }
  }, [isOpen, deviceId]);

  const getStatusColor = (status?: string) => {
    switch (status?.toLowerCase()) {
      case 'connected':
        return 'bg-green-500';
      case 'disconnected':
        return 'bg-red-500';
      case 'connecting':
        return 'bg-yellow-500';
      default:
        return 'bg-gray-500';
    }
  };

  const getStatusIcon = (connected?: boolean) => {
    if (connected) {
      return <CheckCircle className="h-4 w-4 text-green-500" />;
    }
    return <XCircle className="h-4 w-4 text-red-500" />;
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-md max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-green-500"></div>
              <span>Whacenter Device Status</span>
            </div>
            {deviceName && (
              <Badge variant="outline" className="ml-auto">
                {deviceName}
              </Badge>
            )}
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {loading && (
            <div className="flex items-center justify-center p-8">
              <Loader2 className="h-6 w-6 animate-spin mr-2" />
              <span className="text-sm text-gray-600">Checking device status...</span>
            </div>
          )}

          {error && (
            <div className="bg-red-50 border border-red-200 p-3 rounded-md">
              <div className="flex items-center gap-2 mb-2">
                <XCircle className="h-4 w-4 text-red-500" />
                <span className="font-medium text-red-800">Error</span>
              </div>
              <p className="text-red-700 text-sm">{error}</p>
            </div>
          )}

          {status && !loading && (
            <>
              {/* Status Information */}
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-600">Status:</span>
                  <div className="flex items-center gap-2">
                    <div className={`w-2 h-2 rounded-full ${getStatusColor(status.status)}`}></div>
                    <Badge variant={status.connected ? 'default' : 'destructive'}>
                      {status.status?.toUpperCase() || 'UNKNOWN'}
                    </Badge>
                  </div>
                </div>
                
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-600">Connected:</span>
                  <div className="flex items-center gap-2">
                    {getStatusIcon(status.connected)}
                    <Badge variant={status.connected ? 'default' : 'secondary'}>
                      {status.connected ? 'Yes' : 'No'}
                    </Badge>
                  </div>
                </div>
                
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-600">Last Checked:</span>
                  <span className="text-sm text-gray-800">
                    {status.last_checked ? new Date(status.last_checked).toLocaleString() : 'Invalid Date'}
                  </span>
                </div>
              </div>

              {/* QR Code Section */}
              {(() => {
                const qrData = status.qr_code || status.details?.qr_code || status.details?.qr || status.qr;
                const isTimeout = qrData === 'timeout';
                const hasQrData = qrData && !isTimeout;
                
                if (isTimeout) {
                  return (
                    <div className="bg-orange-50 border border-orange-200 p-4 rounded-lg">
                      <div className="flex items-center gap-2 mb-3">
                        <AlertCircle className="h-4 w-4 text-orange-500" />
                        <span className="font-medium text-orange-800">QR Code Status</span>
                      </div>
                      <div className="text-orange-700 text-sm">
                        <p className="font-medium mb-2">Session Timeout</p>
                        <p className="mb-3">The WhatsApp session has expired. Please refresh to generate a new QR code.</p>
                        <Button 
                          onClick={fetchDeviceStatus} 
                          variant="outline" 
                          size="sm" 
                          className="text-orange-700 border-orange-300 hover:bg-orange-100"
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
                    <div className="bg-green-50 border border-green-200 p-4 rounded-lg">
                      <div className="flex items-center gap-2 mb-3">
                        <CheckCircle className="h-4 w-4 text-green-500" />
                        <span className="font-medium text-green-800">QR Code Available</span>
                      </div>
                      <div className="bg-white border rounded-lg p-3">
                        <QRCodeDisplay qrData={qrData} />
                      </div>
                    </div>
                  );
                }
                
                return (
                  <div className="bg-gray-50 border border-gray-200 p-4 rounded-lg">
                    <div className="flex items-center gap-2 mb-3">
                      <AlertCircle className="h-4 w-4 text-gray-500" />
                      <span className="font-medium text-gray-800">QR Code Status</span>
                    </div>
                    <div className="text-gray-700 text-sm">
                      <p className="mb-3">No QR code data available. The device may already be connected or there might be an issue.</p>
                      <Button 
                        onClick={fetchDeviceStatus} 
                        variant="outline" 
                        size="sm"
                      >
                        <RefreshCw className="h-3 w-3 mr-1" />
                        Check Again
                      </Button>
                    </div>
                  </div>
                );
              })()}

              {/* Device Details */}
              {status.details && Object.keys(status.details).length > 0 && (
                <div className="bg-gray-50 border border-gray-200 p-3 rounded-md">
                  <div className="flex items-center gap-2 mb-2">
                    <span className="font-medium text-gray-800">Device Details</span>
                  </div>
                  <div className="space-y-2 text-sm">
                    {Object.entries(status.details)
                      .filter(([key, value]) => key !== 'qr_code' && key !== 'qr' && value !== null && value !== undefined)
                      .map(([key, value]) => (
                      <div key={key} className="flex flex-col gap-1">
                        <span className="text-gray-600 capitalize font-medium">{key.replace(/_/g, ' ')}:</span>
                        <span className="text-gray-800 text-xs bg-white px-2 py-1 rounded border break-all">
                          {typeof value === 'object' ? JSON.stringify(value, null, 2) : String(value)}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </>
          )}

          {/* Action Buttons */}
          <div className="flex gap-2 pt-4">
            <Button 
              onClick={fetchDeviceStatus} 
              variant="outline" 
              className="flex-1"
              disabled={loading}
            >
              <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
            <Button onClick={onClose} className="flex-1">
              Close
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default WhacenterStatusModal;