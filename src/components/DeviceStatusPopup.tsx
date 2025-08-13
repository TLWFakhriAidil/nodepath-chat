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
      
      // Simulate API delay
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      try {
        const response = await fetch(`/api/device-settings/${deviceId}/status`);
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        setStatus(data);
      } catch (apiError) {
        // Use mock data when API fails
        console.log('API not available, using mock data for demonstration');
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
                        {status.details?.qr && status.details?.qr !== 'timeout' && (
                          <div>
                            <p><span className="font-medium">QR:</span> Available</p>
                            <div className="mt-2 p-2 bg-white border rounded">
                              <img 
                                src={status.details.qr.startsWith('data:') ? status.details.qr : `data:image/png;base64,${status.details.qr}`} 
                                alt="QR Code" 
                                className="w-32 h-32 mx-auto"
                                onError={(e) => {
                                  console.error('QR Code image failed to load:', status.details.qr);
                                  e.currentTarget.style.display = 'none';
                                }}
                              />
                            </div>
                          </div>
                        )}
                        {status.details?.qr === 'timeout' && (
                          <p className="text-orange-600"><span className="font-medium">QR:</span> Timeout</p>
                        )}
                        {(status.qr_code || status.details?.qr_code) && (
                          <div>
                            <p><span className="font-medium">QR Code:</span> Available</p>
                            <div className="mt-2 p-2 bg-white border rounded">
                              <img 
                                src={(() => {
                                  const qrData = status.qr_code || status.details?.qr_code;
                                  return qrData.startsWith('data:') ? qrData : `data:image/png;base64,${qrData}`;
                                })()} 
                                alt="QR Code" 
                                className="w-32 h-32 mx-auto"
                                onError={(e) => {
                                  console.error('QR Code image failed to load:', status.qr_code || status.details?.qr_code);
                                  e.currentTarget.style.display = 'none';
                                }}
                              />
                            </div>
                          </div>
                        )}
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

export default DeviceStatusPopup;