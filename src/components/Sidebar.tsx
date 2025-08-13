import { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import {
  LayoutDashboard,
  Workflow,
  MessageSquare,
  FolderOpen,
  BarChart3,
  Settings,
  ChevronLeft,
  ChevronRight,
  Bot,
  User,
  LogOut,
  HelpCircle,
  Zap,
  List,
  Smartphone
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';

interface SidebarProps {
  collapsed: boolean;
  onToggle: () => void;
}

const Sidebar = ({ collapsed, onToggle }: SidebarProps) => {
  const location = useLocation();
  const [notifications] = useState({
    flows: 2,
    messages: 5
  });

  const navigation = [
    {
      name: 'Dashboard',
      href: '/',
      icon: LayoutDashboard,
      current: location.pathname === '/'
    },
    {
      name: 'Flow Builder',
      href: '/flow-builder',
      icon: Workflow,
      current: location.pathname === '/flow-builder',
      badge: notifications.flows
    },
    {
      name: 'Flow Manager',
      href: '/flow-manager',
      icon: List,
      current: location.pathname === '/flow-manager'
    },
    {
      name: 'Test Chat',
      href: '/test-chat',
      icon: MessageSquare,
      current: location.pathname === '/test-chat',
      badge: notifications.messages
    },
    {
      name: 'Analytics',
      href: '/analytics',
      icon: BarChart3,
      current: location.pathname === '/analytics'
    },
    {
      name: 'Device Settings',
      href: '/device-settings',
      icon: Smartphone,
      current: location.pathname === '/device-settings'
    }
  ];

  const quickActions = [
    {
      name: 'New Flow',
      icon: Workflow,
      action: () => console.log('Create new flow')
    },
    {
      name: 'Quick Test',
      icon: Zap,
      action: () => console.log('Quick test')
    }
  ];

  return (
    <div className={`bg-white dark:bg-slate-900 border-r border-slate-200 dark:border-slate-700 transition-all duration-300 ${collapsed ? 'w-16' : 'w-64'} flex flex-col h-full`}>
      {/* Header */}
      <div className="p-4 border-b border-slate-200 dark:border-slate-700">
        <div className="flex items-center justify-between">
          {!collapsed && (
            <div className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                <Bot className="w-5 h-5 text-white" />
              </div>
              <div>
                <h1 className="text-lg font-bold text-slate-900 dark:text-white">
                  ChatBot
                </h1>
                <p className="text-xs text-slate-500 dark:text-slate-400">
                  Builder Pro
                </p>
              </div>
            </div>
          )}
          
          <Button
            variant="ghost"
            size="sm"
            onClick={onToggle}
            className="p-1.5 h-auto"
          >
            {collapsed ? (
              <ChevronRight className="w-4 h-4" />
            ) : (
              <ChevronLeft className="w-4 h-4" />
            )}
          </Button>
        </div>
      </div>

      {/* Navigation */}
      <div className="flex-1 overflow-y-auto py-4">
        <nav className="space-y-1 px-3">
          {navigation.map((item) => {
            const Icon = item.icon;
            return (
              <Link
                key={item.name}
                to={item.href}
                className={`group flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors ${
                  item.current
                    ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-400'
                    : 'text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800'
                }`}
              >
                <Icon className={`flex-shrink-0 w-5 h-5 ${collapsed ? 'mx-auto' : 'mr-3'}`} />
                {!collapsed && (
                  <>
                    <span className="flex-1">{item.name}</span>
                    {item.badge && (
                      <Badge 
                        variant="secondary" 
                        className="ml-auto bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400"
                      >
                        {item.badge}
                      </Badge>
                    )}
                  </>
                )}
              </Link>
            );
          })}
        </nav>

        {!collapsed && (
          <>
            <Separator className="my-4 mx-3" />
            
            {/* Quick Actions */}
            <div className="px-3">
              <p className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-2">
                Quick Actions
              </p>
              <div className="space-y-1">
                {quickActions.map((action) => {
                  const Icon = action.icon;
                  return (
                    <Button
                      key={action.name}
                      variant="ghost"
                      size="sm"
                      onClick={action.action}
                      className="w-full justify-start text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
                    >
                      <Icon className="w-4 h-4 mr-2" />
                      {action.name}
                    </Button>
                  );
                })}
              </div>
            </div>

            <Separator className="my-4 mx-3" />
            
            {/* Status */}
            <div className="px-3">
              <div className="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-3">
                <div className="flex items-center space-x-2">
                  <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
                  <span className="text-xs font-medium text-green-700 dark:text-green-400">
                    System Online
                  </span>
                </div>
                <p className="text-xs text-green-600 dark:text-green-500 mt-1">
                  All services running
                </p>
              </div>
            </div>
          </>
        )}
      </div>

      {/* User Profile */}
      <div className="border-t border-slate-200 dark:border-slate-700 p-3">
        {collapsed ? (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm" className="w-full p-2">
                <Avatar className="w-6 h-6">
                  <AvatarImage src="/placeholder-avatar.jpg" />
                  <AvatarFallback className="bg-blue-100 text-blue-700 text-xs">
                    JD
                  </AvatarFallback>
                </Avatar>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent side="right" align="end" className="w-56">
              <div className="px-2 py-1.5">
                <p className="text-sm font-medium">John Doe</p>
                <p className="text-xs text-slate-500">john@example.com</p>
              </div>
              <DropdownMenuSeparator />
              <DropdownMenuItem>
                <User className="w-4 h-4 mr-2" />
                Profile
              </DropdownMenuItem>
              <DropdownMenuItem>
                <Settings className="w-4 h-4 mr-2" />
                Settings
              </DropdownMenuItem>
              <DropdownMenuItem>
                <HelpCircle className="w-4 h-4 mr-2" />
                Help & Support
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem className="text-red-600">
                <LogOut className="w-4 h-4 mr-2" />
                Sign Out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        ) : (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="w-full justify-start p-2 h-auto">
                <div className="flex items-center space-x-3">
                  <Avatar className="w-8 h-8">
                    <AvatarImage src="/placeholder-avatar.jpg" />
                    <AvatarFallback className="bg-blue-100 text-blue-700">
                      JD
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1 text-left">
                    <p className="text-sm font-medium text-slate-900 dark:text-white">
                      John Doe
                    </p>
                    <p className="text-xs text-slate-500 dark:text-slate-400">
                      john@example.com
                    </p>
                  </div>
                </div>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent side="right" align="end" className="w-56">
              <DropdownMenuItem>
                <User className="w-4 h-4 mr-2" />
                Profile
              </DropdownMenuItem>
              <DropdownMenuItem>
                <Settings className="w-4 h-4 mr-2" />
                Settings
              </DropdownMenuItem>
              <DropdownMenuItem>
                <HelpCircle className="w-4 h-4 mr-2" />
                Help & Support
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem className="text-red-600">
                <LogOut className="w-4 h-4 mr-2" />
                Sign Out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </div>
    </div>
  );
};

export default Sidebar;