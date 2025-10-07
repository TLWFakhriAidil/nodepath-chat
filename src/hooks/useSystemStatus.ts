import { useState, useEffect } from 'react';

interface User {
  id: string;
  email: string;
  full_name: string;
  status: string;
  expired?: string;
  is_active: boolean;
}

interface SystemStatus {
  isOnline: boolean;
  userType: 'pro' | 'trial' | 'expired';
  displayText: string;
  statusColor: 'green' | 'yellow' | 'red';
}

export const useSystemStatus = (): SystemStatus => {
  const [systemStatus, setSystemStatus] = useState<SystemStatus>({
    isOnline: true,
    userType: 'trial',
    displayText: 'System Online (Trial)',
    statusColor: 'green'
  });

  const checkSystemStatus = async () => {
    try {
      // Fetch user profile to get status and expiration
      const response = await fetch('/api/profile/', {
        credentials: 'include',
      });

      if (!response.ok) {
        // If can't fetch profile, default to offline
        setSystemStatus({
          isOnline: false,
          userType: 'expired',
          displayText: 'System Offline (Connection Error)',
          statusColor: 'red'
        });
        return;
      }

      const result = await response.json();
      if (result.success && result.data) {
        const user: User = result.data;
        const status = determineSystemStatus(user);
        setSystemStatus(status);
      } else {
        setSystemStatus({
          isOnline: false,
          userType: 'expired',
          displayText: 'System Offline (Auth Error)',
          statusColor: 'red'
        });
      }
    } catch (error) {
      console.error('Error checking system status:', error);
      setSystemStatus({
        isOnline: false,
        userType: 'expired',
        displayText: 'System Offline (Network Error)',
        statusColor: 'red'
      });
    }
  };

  const determineSystemStatus = (user: User): SystemStatus => {
    // Check if user has expired date and if it's passed
    if (user.expired) {
      try {
        // Parse the expired date - handle format "2025-10-13 08:27:12"
        const expiredDate = new Date(user.expired);
        const now = new Date();
        
        // Check if the date is valid and if current time has passed expiration
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

    // If not expired, check user status
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
      // Any other status (like 'expired', 'inactive', etc.)
      return {
        isOnline: false,
        userType: 'expired',
        displayText: 'System Offline (Expired)',
        statusColor: 'red'
      };
    }
  };

  useEffect(() => {
    checkSystemStatus();
    
    // Check status every 5 minutes
    const interval = setInterval(checkSystemStatus, 5 * 60 * 1000);
    
    return () => clearInterval(interval);
  }, []);

  return systemStatus;
};