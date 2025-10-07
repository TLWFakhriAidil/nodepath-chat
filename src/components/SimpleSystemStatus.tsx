import React, { useState, useEffect } from 'react';

// Clean, simple system status component - NO COMPLEX HOOKS
const SimpleSystemStatus: React.FC = () => {
  const [status, setStatus] = useState({
    text: 'System Online (Trial)',
    color: 'yellow',
    isLoading: true
  });

  useEffect(() => {
    console.log('🎯 SimpleSystemStatus component loaded');
    
    const checkUserStatus = async () => {
      try {
        console.log('📡 Fetching /api/profile/status...');
        
        const response = await fetch('/api/profile/status', {
          credentials: 'include'
        });
        
        console.log('📡 Response status:', response.status);
        
        if (response.ok) {
          const data = await response.json();
          console.log('📋 Response data:', data);
          
          if (data.success && data.data) {
            const userStatus = data.data.status?.toLowerCase() || '';
            const expired = data.data.expired;
            
            console.log('👤 User status:', userStatus);
            console.log('📅 Expired date:', expired);
            
            // Check if expired
            if (expired) {
              const expiredDate = new Date(expired);
              const now = new Date();
              
              if (now > expiredDate) {
                console.log('❌ User is expired');
                setStatus({
                  text: 'System Offline (Expired)',
                  color: 'red',
                  isLoading: false
                });
                return;
              }
            }
            
            // Check status
            if (userStatus === 'pro') {
              console.log('✅ User is Pro');
              setStatus({
                text: 'System Online (Pro)',
                color: 'green',
                isLoading: false
              });
            } else if (userStatus === 'trial') {
              console.log('✅ User is Trial');
              setStatus({
                text: 'System Online (Trial)',
                color: 'yellow',
                isLoading: false
              });
            } else {
              console.log('❌ Unknown status:', userStatus);
              setStatus({
                text: 'System Offline (Unknown)',
                color: 'red',
                isLoading: false
              });
            }
          } else {
            console.log('❌ Invalid response format');
            setStatus({
              text: 'System Offline (Data Error)',
              color: 'red',
              isLoading: false
            });
          }
        } else {
          console.log('❌ API call failed:', response.status);
          setStatus({
            text: `System Offline (API ${response.status})`,
            color: 'red',
            isLoading: false
          });
        }
      } catch (error) {
        console.error('❌ Network error:', error);
        setStatus({
          text: 'System Offline (Network)',
          color: 'red',
          isLoading: false
        });
      }
    };
    
    checkUserStatus();
  }, []);

  const getColors = () => {
    switch (status.color) {
      case 'green':
        return {
          bg: 'bg-green-50 dark:bg-green-900/20',
          border: 'border-green-200 dark:border-green-800',
          dot: 'bg-green-500',
          text: 'text-green-700 dark:text-green-400',
          subtext: 'text-green-600 dark:text-green-500'
        };
      case 'yellow':
        return {
          bg: 'bg-yellow-50 dark:bg-yellow-900/20',
          border: 'border-yellow-200 dark:border-yellow-800',
          dot: 'bg-yellow-500',
          text: 'text-yellow-700 dark:text-yellow-400',
          subtext: 'text-yellow-600 dark:text-yellow-500'
        };
      default:
        return {
          bg: 'bg-red-50 dark:bg-red-900/20',
          border: 'border-red-200 dark:border-red-800',
          dot: 'bg-red-500',
          text: 'text-red-700 dark:text-red-400',
          subtext: 'text-red-600 dark:text-red-500'
        };
    }
  };

  const colors = getColors();

  return (
    <div className={`${colors.bg} ${colors.border} border rounded-lg p-3`}>
      <div className="flex items-center space-x-2">
        <div className={`w-2 h-2 rounded-full ${colors.dot} ${status.isLoading ? 'animate-pulse' : status.color !== 'red' ? 'animate-pulse' : ''}`} />
        <span className={`text-xs font-medium ${colors.text}`}>
          {status.isLoading ? 'Checking status...' : status.color === 'red' ? 'Service unavailable' : 'All services running'}
        </span>
      </div>
    </div>
  );
};

export default SimpleSystemStatus;