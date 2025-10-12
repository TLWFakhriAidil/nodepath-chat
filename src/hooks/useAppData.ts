import { useState, useEffect, useCallback } from 'react';

/**
 * Unified application data interface
 * Consolidates user profile, device status, and system status into one call
 */
interface AppData {
  user: {
    id: string;
    email: string;
    full_name: string;
    status: string;
    expired?: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
    last_login?: string;
  };
  devices: {
    has_devices: boolean;
    device_count: number;
    device_ids: string[];
  };
  system: {
    isOnline: boolean;
    userType: 'pro' | 'trial' | 'expired';
    displayText: string;
    statusColor: 'green' | 'yellow' | 'red';
  };
}

interface UseAppDataReturn {
  data: AppData | null;
  isLoading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
  // Legacy compatibility props
  user: AppData['user'] | null;
  deviceStatus: AppData['devices'] | null;
  systemStatus: AppData['system'];
  hasDevices: boolean;
  has_devices: boolean;
  device_ids: string[];
}

/**
 * Cache management for app data
 */
class AppDataCache {
  private data: AppData | null = null;
  private lastFetch: number = 0;
  private readonly cacheDuration = 5 * 60 * 1000; // 5 minutes

  set(data: AppData) {
    this.data = data;
    this.lastFetch = Date.now();
  }

  get(): AppData | null {
    if (!this.data || this.isExpired()) {
      return null;
    }
    return this.data;
  }

  isExpired(): boolean {
    return Date.now() - this.lastFetch > this.cacheDuration;
  }

  clear() {
    this.data = null;
    this.lastFetch = 0;
  }
}

// Global cache instance
const appDataCache = new AppDataCache();

/**
 * Optimized hook that consolidates all sidebar-related data fetching
 * Replaces useDevice, useSystemStatus, and profile calls with a single optimized request
 */
export const useAppData = (): UseAppDataReturn => {
  const [data, setData] = useState<AppData | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  /**
   * Determine system status from user data
   */
  const determineSystemStatus = useCallback((user: AppData['user']): AppData['system'] => {
    // Check if user has expired date and if it's passed
    if (user.expired) {
      try {
        const expiredDate = new Date(user.expired);
        const now = new Date();
        
        if (!isNaN(expiredDate.getTime()) && now > expiredDate) {
          return {
            isOnline: false,
            userType: 'expired',
            displayText: 'System Offline (Expired)',
            statusColor: 'red'
          };
        }
      } catch (error) {
        console.error('Error parsing expired date:', error);
      }
    }

    // Check user status
    const userStatus = user.status?.toLowerCase() || '';
    
    if (userStatus === 'pro') {
      return {
        isOnline: true,
        userType: 'pro',
        displayText: 'System Online (Pro)',
        statusColor: 'green'
      };
    } else if (userStatus === 'trial') {
      return {
        isOnline: true,
        userType: 'trial',
        displayText: 'System Online (Trial)',
        statusColor: 'yellow'
      };
    } else {
      return {
        isOnline: false,
        userType: 'expired',
        displayText: 'System Offline (Expired)',
        statusColor: 'red'
      };
    }
  }, []);

  /**
   * Unified data fetching function
   * Single API call that returns all required sidebar data
   */
  const fetchAppData = useCallback(async (): Promise<void> => {
    // Check cache first
    const cachedData = appDataCache.get();
    if (cachedData) {
      setData(cachedData);
      setError(null);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      console.log('🚀 Fetching consolidated app data...');
      
      // Single optimized API call that returns user + device data
      const response = await fetch('/api/app/data', {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`API call failed with status: ${response.status}`);
      }

      const result = await response.json();
      
      if (!result.success) {
        throw new Error(result.error || 'Failed to fetch app data');
      }

      // Construct unified app data
      const appData: AppData = {
        user: result.data.user,
        devices: {
          has_devices: result.data.has_devices,
          device_count: result.data.device_count,
          device_ids: result.data.device_ids || [],
        },
        system: determineSystemStatus(result.data.user)
      };

      // Cache the result
      appDataCache.set(appData);
      setData(appData);

      console.log('✅ Consolidated app data loaded successfully');
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      console.error('❌ Failed to fetch app data:', errorMessage);
      setError(errorMessage);
      
      // Clear cache on error
      appDataCache.clear();
    } finally {
      setIsLoading(false);
    }
  }, [determineSystemStatus]);

  /**
   * Refetch function for manual refresh
   */
  const refetch = useCallback(async (): Promise<void> => {
    appDataCache.clear(); // Force cache miss
    await fetchAppData();
  }, [fetchAppData]);

  // Initial data fetch
  useEffect(() => {
    fetchAppData();
  }, [fetchAppData]);

  // Auto-refresh every 5 minutes
  useEffect(() => {
    const interval = setInterval(() => {
      if (appDataCache.isExpired()) {
        fetchAppData();
      }
    }, 60 * 1000); // Check every minute

    return () => clearInterval(interval);
  }, [fetchAppData]);

  // Legacy compatibility properties
  const hasDevices = data?.devices.has_devices ?? false;

  return {
    data,
    isLoading,
    error,
    refetch,
    // Legacy compatibility
    user: data?.user || null,
    deviceStatus: data?.devices || null,
    systemStatus: data?.system || {
      isOnline: false,
      userType: 'expired',
      displayText: 'Loading...',
      statusColor: 'red'
    },
    hasDevices,
    has_devices: hasDevices,
    device_ids: data?.devices.device_ids || [],
  };
};

/**
 * Hook for components that only need device data
 * Uses the same cache as useAppData for efficiency
 */
export const useOptimizedDevice = () => {
  const { deviceStatus, has_devices, device_ids, isLoading, refetch } = useAppData();
  
  return {
    deviceStatus,
    isLoading,
    hasDevices: has_devices,
    has_devices,
    device_ids,
    checkDeviceStatus: refetch,
    refreshDeviceStatus: refetch,
  };
};

/**
 * Hook for components that only need system status
 * Uses the same cache as useAppData for efficiency
 */
export const useOptimizedSystemStatus = () => {
  const { systemStatus } = useAppData();
  return systemStatus;
};