import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useAuth } from './AuthContext';

/**
 * Interface for device status information
 */
interface DeviceStatus {
  has_devices: boolean;
  device_count: number;
  device_ids: string[];
}

/**
 * Interface for device context type
 */
interface DeviceContextType {
  deviceStatus: DeviceStatus | null;
  isLoading: boolean;
  hasDevices: boolean;
  has_devices: boolean; // Alias for hasDevices for consistency
  device_ids: string[]; // Direct access to device IDs
  checkDeviceStatus: () => Promise<void>;
  refreshDeviceStatus: () => Promise<void>;
}

// Create the device context
const DeviceContext = createContext<DeviceContextType | undefined>(undefined);

/**
 * Custom hook to use the device context
 */
export const useDevice = () => {
  const context = useContext(DeviceContext);
  if (context === undefined) {
    throw new Error('useDevice must be used within a DeviceProvider');
  }
  return context;
};

/**
 * Device provider component props
 */
interface DeviceProviderProps {
  children: ReactNode;
}

/**
 * Device provider component that manages device status state
 */
export const DeviceProvider: React.FC<DeviceProviderProps> = ({ children }) => {
  const [deviceStatus, setDeviceStatus] = useState<DeviceStatus | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const { isAuthenticated, user } = useAuth();

  // Computed property for easier access
  const hasDevices = deviceStatus?.has_devices ?? false;

  /**
   * Check device status from the backend API
   */
  const checkDeviceStatus = async () => {
    if (!isAuthenticated || !user) {
      setDeviceStatus(null);
      return;
    }

    try {
      setIsLoading(true);
      const response = await fetch('/api/auth/device-status', {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const result = await response.json();
        if (result.success) {
          setDeviceStatus({
            has_devices: result.has_devices,
            device_count: result.device_count,
            device_ids: result.device_ids || [],
          });
        } else {
          console.error('Failed to fetch device status:', result.error);
          setDeviceStatus({ has_devices: false, device_count: 0, device_ids: [] });
        }
      } else {
        console.error('Device status request failed:', response.status);
        setDeviceStatus({ has_devices: false, device_count: 0, device_ids: [] });
      }
    } catch (error) {
      console.error('Error checking device status:', error);
      setDeviceStatus({ has_devices: false, device_count: 0, device_ids: [] });
    } finally {
      setIsLoading(false);
    }
  };

  /**
   * Refresh device status (alias for checkDeviceStatus)
   */
  const refreshDeviceStatus = checkDeviceStatus;

  // Check device status when authentication state changes
  useEffect(() => {
    if (isAuthenticated && user) {
      checkDeviceStatus();
    } else {
      setDeviceStatus(null);
    }
  }, [isAuthenticated, user]);

  const value: DeviceContextType = {
    deviceStatus,
    isLoading,
    hasDevices,
    has_devices: hasDevices, // Alias for consistency
    device_ids: deviceStatus?.device_ids || [], // Direct access to device IDs
    checkDeviceStatus,
    refreshDeviceStatus,
  };

  return (
    <DeviceContext.Provider value={value}>
      {children}
    </DeviceContext.Provider>
  );
};