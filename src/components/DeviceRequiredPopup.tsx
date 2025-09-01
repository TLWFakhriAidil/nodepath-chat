import React from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { AlertTriangle, Smartphone } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

interface DeviceRequiredPopupProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  featureName?: string;
}

/**
 * Popup component that notifies users when they need to configure devices
 * to access certain features. Provides navigation to device settings.
 */
export const DeviceRequiredPopup: React.FC<DeviceRequiredPopupProps> = ({
  open,
  onOpenChange,
  featureName = 'this feature'
}) => {
  const navigate = useNavigate();

  /**
   * Handle navigation to device settings page
   * Closes the popup and redirects user to configure devices
   */
  const handleGoToDeviceSettings = () => {
    onOpenChange(false);
    navigate('/device-settings');
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-3 mb-2">
            <div className="w-12 h-12 bg-orange-100 dark:bg-orange-900/20 rounded-full flex items-center justify-center">
              <AlertTriangle className="w-6 h-6 text-orange-600 dark:text-orange-400" />
            </div>
            <div>
              <DialogTitle className="text-lg font-semibold">
                Device Configuration Required
              </DialogTitle>
            </div>
          </div>
          <DialogDescription className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
            To access {featureName}, you need to configure at least one device in your account. 
            Devices are required to manage chatbot flows, analytics, and messaging features.
          </DialogDescription>
        </DialogHeader>
        
        <div className="flex flex-col gap-3 mt-4">
          <Button 
            onClick={handleGoToDeviceSettings}
            className="w-full flex items-center gap-2"
          >
            <Smartphone className="w-4 h-4" />
            Go to Device Settings
          </Button>
          
          <Button 
            variant="outline" 
            onClick={() => onOpenChange(false)}
            className="w-full"
          >
            Cancel
          </Button>
        </div>
        
        <div className="mt-4 p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
          <p className="text-xs text-blue-700 dark:text-blue-300">
            <strong>Tip:</strong> Once you configure your first device, you'll have access to all chatbot features including flow building, analytics, and message management.
          </p>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default DeviceRequiredPopup;