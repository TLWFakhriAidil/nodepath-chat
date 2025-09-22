import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDevice } from '@/contexts/DeviceContext';
import { useToast } from '@/hooks/use-toast';

interface DeviceRequiredWrapperProps {
  children: React.ReactNode;
}

export default function DeviceRequiredWrapper({ children }: DeviceRequiredWrapperProps) {
  const { has_devices } = useDevice();
  const navigate = useNavigate();
  const { toast } = useToast();

  useEffect(() => {
    if (!has_devices) {
      // Show error modal
      toast({
        title: "Device Required",
        description: "Please add a device first to access this feature",
        variant: "destructive",
      });
      
      // Redirect to device settings after a short delay
      setTimeout(() => {
        navigate('/device-settings');
      }, 1500);
    }
  }, [has_devices, navigate, toast]);

  // Don't render children if no devices
  if (!has_devices) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <h2 className="text-2xl font-bold mb-2">Device Required</h2>
          <p className="text-muted-foreground">Redirecting to Device Settings...</p>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
