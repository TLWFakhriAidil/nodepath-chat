import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Loader2, RefreshCw, CheckCircle, XCircle, AlertCircle, QrCode } from 'lucide-react';
import { toast } from 'sonner';

/**
 * Interface for WAHA device status response
 */
interface WahaDeviceStatus {
  provider?: string;
  status?: string;
  image?: string; // Base64 QR code image
}

/**
 * Props for WAHA Status Modal component
 */
interface WahaStatusModalProps {
  isOpen: boolean;
  onClose: () => void;
  deviceId: string;
  deviceName?: string;
}

/**
 * WAHA Status Modal Component
 * Displays device status and QR code for WAHA provider authentication
 */
const WahaStatusModal: React.FC<WahaStatusModalProps> = ({
  isOpen,
  onClose,
  deviceId,
  deviceName
}) => {
  const [status, setStatus] = useState<WahaDeviceStatus | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  /**
   * Fetch WAHA device status from backend API
   */
  const fetchDeviceStatus = async () => {
    setLoading(true);
    setError(null);
    
    try {
      console.log('Fetching WAHA device status for device ID:', deviceId);
      const response = await fetch(`/api/device-settings/${deviceId}/waha-status`);
      const data = await response.json();
      
      console.log('WAHA status response:', data);
      
      if (response.ok && data) {
        setStatus(data);
      } else {
        setError(data.error || 'Failed to fetch WAHA device status');
      }
    } catch (err) {
      console.error('Error fetching WAHA device status:', err);
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

  /**
   * Get status color based on WAHA status
   */
  const getStatusColor = (status?: string) => {
    switch (status?.toUpperCase()) {
      case 'WORKING':
        return 'bg-green-500';
      case 'SCAN_QR_CODE':
        return 'bg-yellow-500';
      case 'STOPPED':
        return 'bg-red-500';
      case 'STARTING':
        return 'bg-blue-500';
      default:
        return 'bg-gray-500';
    }
  };

  /**
   * Get status icon based on WAHA status
   */
  const getStatusIcon = (status?: string) => {
    switch (status?.toUpperCase()) {
      case 'WORKING':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'SCAN_QR_CODE':
        return <QrCode className="h-4 w-4 text-yellow-500" />;
      case 'STOPPED':
        return <XCircle className="h-4 w-4 text-red-500" />;
      case 'STARTING':
        return <Loader2 className="h-4 w-4 text-blue-500 animate-spin" />;
      default:
        return <AlertCircle className="h-4 w-4 text-gray-500" />;
    }
  };

  /**
   * Get user-friendly status message
   */
  const getStatusMessage = (status?: string) => {
    switch (status?.toUpperCase()) {
      case 'WORKING':
        return 'Device is connected and working properly';
      case 'SCAN_QR_CODE':
        return 'Please scan the QR code below to authenticate';
      case 'STOPPED':
        return 'Device session is stopped. Click refresh to restart.';
      case 'STARTING':
        return 'Device session is starting up...';
      default:
        return 'Unknown device status';
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-purple-500"></div>
              <span>WAHA Device Status</span>
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
              <span className="text-sm text-gray-600">Checking WAHA device status...</span>
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
              <div className="grid grid-cols-1 gap-4">
                <div>
                  <label className="text-sm font-medium text-gray-600">Status:</label>
                  <div className="flex items-center gap-2 mt-1">
                    <div className={`w-2 h-2 rounded-full ${getStatusColor(status.status)}`}></div>
                    {getStatusIcon(status.status)}
                    <Badge variant={status.status === 'WORKING' ? 'default' : 'secondary'}>
                      {status.status?.toUpperCase() || 'UNKNOWN'}
                    </Badge>
                  </div>
                </div>
                
                <div>
                  <label className="text-sm font-medium text-gray-600">Provider:</label>
                  <div className="mt-1">
                    <Badge variant="outline" className="bg-purple-50 text-purple-700 border-purple-200">
                      WAHA
                    </Badge>
                  </div>
                </div>
                
                <div>
                  <label className="text-sm font-medium text-gray-600">Message:</label>
                  <div className="mt-1">
                    <span className="text-sm text-gray-800">
                      {getStatusMessage(status.status)}
                    </span>
                  </div>
                </div>
              </div>

              {/* QR Code Section */}
              {status.status === 'SCAN_QR_CODE' && status.image && (
                <div className="bg-yellow-50 border border-yellow-200 p-4 rounded-lg">
                  <div className="flex items-center justify-center mb-3">
                    <QrCode className="w-6 h-6 text-yellow-600 mr-2" />
                    <span className="font-medium text-yellow-800">WhatsApp QR Code</span>
                  </div>
                  <p className="text-sm text-yellow-700 mb-4 text-center">
                    Scan this QR code with your WhatsApp mobile app:
                  </p>
                  <div className="flex justify-center mb-4">
                    <div className="bg-white p-4 rounded-lg shadow-sm border">
                      <img 
                        src={status.image} 
                        alt="WAHA QR Code" 
                        className="w-48 h-48 object-contain"
                        onError={(e) => {
                          console.error('QR Code image failed to load');
                          e.currentTarget.style.display = 'none';
                        }}
                      />
                    </div>
                  </div>
                  <p className="text-xs text-yellow-600 text-center">
                    Open WhatsApp → Settings → Linked Devices → Link a Device
                  </p>
                </div>
              )}

              {/* Success Message */}
              {status.status === 'WORKING' && (
                <div className="bg-green-50 border border-green-200 p-4 rounded-lg">
                  <div className="flex items-center justify-center mb-2">
                    <CheckCircle className="w-6 h-6 text-green-600 mr-2" />
                    <span className="font-medium text-green-800">Device Connected</span>
                  </div>
                  <p className="text-sm text-green-700 text-center">
                    Your WAHA device is successfully connected and ready to send/receive messages.
                  </p>
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

export default WahaStatusModal;