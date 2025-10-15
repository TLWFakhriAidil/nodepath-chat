import React from 'react';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/contexts/AuthContext';
import { LogOut, User } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useOptimizedSystemStatus } from '@/hooks/useAppData';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface TopBarProps {
  sidebarOpen: boolean;
  setSidebarOpen: (open: boolean) => void;
}

/**
 * TopBar component that displays the application header with user information and logout functionality
 * Integrates with the authentication context to provide user management features
 */
const TopBar: React.FC<TopBarProps> = ({ sidebarOpen, setSidebarOpen }) => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const systemStatus = useOptimizedSystemStatus();

  // Convert optimized system status to TopBar format
  const topBarStatus = {
    text: systemStatus.displayText,
    color: systemStatus.statusColor,
    isLoading: false
  };

  /**
   * Handle user logout
   * Calls the logout function from the authentication context
   */
  const handleLogout = async () => {
    try {
      await logout();
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  /**
   * Handle profile navigation
   */
  const handleProfileClick = () => {
    navigate('/profile');
  };

  return (
    <div className="sticky top-0 z-30 bg-white/80 dark:bg-slate-900/80 backdrop-blur-xl border-b border-slate-200/50 dark:border-slate-700/50">
      <div className="flex h-16 items-center justify-between px-6">
        <div className="flex items-center space-x-4">
          <div>
            <h1 className="text-lg font-semibold text-slate-900 dark:text-white">
              ChatBot Builder
            </h1>
            <p className="text-sm text-slate-500 dark:text-slate-400">
              Build intelligent conversational flows
            </p>
          </div>
        </div>
        
        <div className="flex items-center space-x-4">
          <div className="hidden md:flex items-center space-x-2 bg-slate-100/50 dark:bg-slate-800/50 rounded-lg px-3 py-1.5">
            <div className={`w-2 h-2 rounded-full ${
              topBarStatus.color === 'green' ? 'bg-green-500 animate-pulse' :
              topBarStatus.color === 'yellow' ? 'bg-yellow-500 animate-pulse' :
              'bg-red-500'
            }`} />
            <span className="text-sm font-bold text-slate-600 dark:text-slate-300">
              {topBarStatus.text}
            </span>
          </div>
          
          {/* User dropdown menu */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="relative h-8 w-8 rounded-full">
                <div className="w-8 h-8 bg-gradient-to-r from-blue-600 to-purple-600 rounded-full flex items-center justify-center">
                  <span className="text-sm font-medium text-white">
                    {user?.full_name ? user.full_name.charAt(0).toUpperCase() : 'U'}
                  </span>
                </div>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-56" align="end" forceMount>
              <DropdownMenuLabel className="font-normal">
                <div className="flex flex-col space-y-1">
                  <p className="text-sm font-medium leading-none">
                    {user?.full_name || 'User'}
                  </p>
                  <p className="text-xs leading-none text-muted-foreground">
                    {user?.email || 'user@example.com'}
                  </p>
                </div>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem className="cursor-pointer" onClick={handleProfileClick}>
                <User className="mr-2 h-4 w-4" />
                <span>Profile</span>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem 
                className="cursor-pointer text-red-600 focus:text-red-600"
                onClick={handleLogout}
              >
                <LogOut className="mr-2 h-4 w-4" />
                <span>Log out</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </div>
  );
};

export default TopBar;