import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Loader2, RefreshCw, CheckCircle, XCircle, AlertCircle, ExternalLink } from 'lucide-react';
import { toast } from 'sonner';

interface DeviceStatus {
  device_id?: string;
  provider?: string;
  connected?: boolean;
  status?: string;
  last_checked?: string;
  qr?: string;
  details?: any;
}

interface WablasStatusModalProps {
  isOpen: boolean;
  onClose: () => void;
  deviceId: string;
  deviceName?: string;
}

const WablasStatusModal: React.FC<WablasStatusModalProps> = ({
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

  const handleOpenQRGenerator = (url: string) => {
    window.open(url, '_blank');
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-blue-500"></div>
              <span>Wablas Device Status</span>
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
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium text-gray-600">Status:</label>
                  <div className="flex items-center gap-2 mt-1">
                    <div className={`w-2 h-2 rounded-full ${getStatusColor(status.status)}`}></div>
                    <Badge variant={status.connected ? 'default' : 'destructive'}>
                      {status.status?.toUpperCase() || 'UNKNOWN'}
                    </Badge>
                  </div>
                </div>
                
                <div>
                  <label className="text-sm font-medium text-gray-600">Provider:</label>
                  <div className="mt-1">
                    <Badge variant="outline">Wablas</Badge>
                  </div>
                </div>
                
                <div>
                  <label className="text-sm font-medium text-gray-600">Connected:</label>
                  <div className="flex items-center gap-2 mt-1">
                    {getStatusIcon(status.connected)}
                    <Badge variant={status.connected ? 'default' : 'secondary'}>
                      {status.connected ? 'Yes' : 'No'}
                    </Badge>
                  </div>
                </div>
                
                <div>
                  <label className="text-sm font-medium text-gray-600">Last Checked:</label>
                  <div className="mt-1">
                    <span className="text-sm text-gray-800">
                      {status.last_checked ? new Date(status.last_checked).toLocaleString() : 'Invalid Date'}
                    </span>
                  </div>
                </div>
              </div>

              {/* QR Code URL Section */}
              {status.qr && (
                <div className="bg-blue-50 border border-blue-200 p-4 rounded-lg">
                  <div className="flex items-center justify-center mb-3">
                    <ExternalLink className="w-6 h-6 text-blue-500 mr-2" />
                    <span className="font-medium text-blue-800">WhatsApp QR Code</span>
                  </div>
                  <p className="text-sm text-blue-700 mb-4 text-center">
                    Click the button below to open the QR code generator:
                  </p>
                  <div className="text-center">
                    <Button 
                      onClick={() => handleOpenQRGenerator(status.qr!)}
                      className="bg-blue-600 hover:bg-blue-700 text-white w-full"
                    >
                      <ExternalLink className="w-4 h-4 mr-2" />
                      Open QR Generator
                    </Button>
                  </div>
                  <p className="text-xs text-blue-600 mt-3 text-center">
                    This will open the Wablas QR code generator in a new tab
                  </p>
                  
                  {/* Show the URL for reference */}
                  <div className="mt-3 p-2 bg-blue-100 rounded text-xs text-blue-800 break-all">
                    <strong>URL:</strong> {status.qr}
                  </div>
                </div>
              )}

              {/* Device Details */}
              {status.details && Object.keys(status.details).length > 0 && (
                <div className="bg-gray-50 border border-gray-200 p-3 rounded-md">
                  <div className="flex items-center gap-2 mb-2">
                    <span className="font-medium text-gray-800">Device Details</span>
                  </div>
                  <div className="space-y-1 text-sm">
                    {Object.entries(status.details).map(([key, value]) => (
                      <div key={key} className="flex justify-between">
                        <span className="text-gray-600 capitalize">{key.replace(/_/g, ' ')}:</span>
                        <span className="text-gray-800 font-mono text-xs">
                          {typeof value === 'object' ? JSON.stringify(value) : String(value)}
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

export default WablasStatusModal;