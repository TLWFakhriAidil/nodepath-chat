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
    displayText: 'System Online (Trial) - Hook Active',
    statusColor: 'yellow'
  });

  const checkSystemStatus = async () => {
    try {
      console.log('🔍 Attempting to fetch /api/profile/...');
      
      // Fetch user profile to get status and expiration
      const response = await fetch('/api/profile/', {
        credentials: 'include',
      });

      console.log('📡 API Response status:', response.status);
      console.log('📡 API Response headers:', response.headers);

      if (!response.ok) {
        console.error('❌ API call failed with status:', response.status);
        // If can't fetch profile, default to offline
        setSystemStatus({
          isOnline: false,
          userType: 'expired',
          displayText: `System Offline (API Error ${response.status})`,
          statusColor: 'red'
        });
        return;
      }

      const result = await response.json();
      console.log('📋 API Response data:', result);
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
    console.log('🔍 Debug user data:', user); // Debug log
    
    // Check if user has expired date and if it's passed
    if (user.expired) {
      try {
        // Parse the expired date - handle format "2025-10-13 08:27:12"
        const expiredDate = new Date(user.expired);
        const now = new Date();
        
        console.log('📅 Expired date:', expiredDate);
        console.log('📅 Current time:', now);
        console.log('⏰ Is expired?', now > expiredDate);
        
        // Check if the date is valid and if current time has passed expiration
        if (!isNaN(expiredDate.getTime()) && now > expiredDate) {
          console.log('❌ User is expired');
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
    console.log('👤 User status:', userStatus);
    
    if (userStatus === 'pro') {
      console.log('✅ User is Pro');
      return {
        isOnline: true,
        userType: 'pro',
        displayText: 'System Online (Pro)',
        statusColor: 'green'
      };
    } else if (userStatus === 'trial') {
      console.log('✅ User is Trial');
      return {
        isOnline: true,
        userType: 'trial',
        displayText: 'System Online (Trial)',
        statusColor: 'yellow'
      };
    } else {
      console.log('❌ User status unrecognized or inactive');
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