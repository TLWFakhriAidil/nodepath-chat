import { Toaster } from '@/components/ui/toaster';
import { Toaster as Sonner } from '@/components/ui/sonner';
import { TooltipProvider } from '@/components/ui/tooltip';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { useState } from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { 
  Bot, 
  Workflow, 
  MessageSquare, 
  BarChart3, 
  Menu,
  X,
  Home,
  Zap,
  Smartphone,
  User,
  CreditCard
} from 'lucide-react';

// Import pages
import Dashboard from './pages/Dashboard';
import FlowBuilder from './pages/FlowBuilder';
import FlowManager from './pages/FlowManager';
import Analytics from './pages/AnalyticsNew';
import DeviceSettings from './pages/DeviceSettings';
import SetStage from './pages/SetStage';
import WhatsAppBot from './pages/WhatsAppBot';
import Profile from './pages/Profile';
import Billings from './pages/Billings';
import Login from './pages/Login';
import Register from './pages/Register';
import { AuthProvider } from './contexts/AuthContext';
import { DeviceProvider, useDevice } from './contexts/DeviceContext';
import ProtectedRoute from './components/ProtectedRoute';
import TopBar from './components/TopBar';
import DeviceRequiredWrapper from './components/DeviceRequiredWrapper';


const queryClient = new QueryClient();

/**
 * Main application content component that uses device context
 * Handles navigation restrictions based on device ownership
 */
function AppContent() {
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const { has_devices } = useDevice();
  
  // DEBUG: App component is rendering
  console.log('🎯 APP CONTENT: AppContent component rendered');
  console.log('🎯 APP CONTENT: has_devices =', has_devices);

  // Navigation items with device restrictions
  const navigation = [
    { name: 'Dashboard', href: '/', icon: Home, current: true, requiresDevice: false },
    { name: 'Device Settings', href: '/device-settings', icon: Smartphone, current: false, requiresDevice: false },
    { name: 'Profile', href: '/profile', icon: User, current: false, requiresDevice: false },
    { name: 'Billings', href: '/billings', icon: CreditCard, current: false, requiresDevice: false },
    { name: 'Set Stage', href: '/set-stage', icon: Zap, current: false, requiresDevice: true },
    { name: 'Flow Builder', href: '/flow-builder', icon: Workflow, current: false, requiresDevice: true },
    { name: 'Flow Manager', href: '/flow-manager', icon: Workflow, current: false, requiresDevice: true },
    { name: 'Chatbot AI', href: '/chatbot-ai', icon: Bot, current: false, requiresDevice: true },
    { name: 'WhatsApp Bot', href: '/whatsapp-bot', icon: MessageSquare, current: false, requiresDevice: true },
  ];

  // Filter navigation based on device ownership
  const filteredNavigation = navigation.filter(item => 
    !item.requiresDevice || has_devices
  );

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-100 dark:from-slate-900 dark:via-slate-800 dark:to-slate-900">
      {/* Sidebar */}
      <div className={cn(
        "fixed inset-y-0 left-0 z-50 w-64 bg-white/80 dark:bg-slate-900/80 backdrop-blur-xl border-r border-slate-200/50 dark:border-slate-700/50 transition-transform duration-300 ease-in-out",
        sidebarOpen ? "translate-x-0" : "-translate-x-full"
      )}>
        <div className="flex h-16 items-center justify-between px-6 border-b border-slate-200/50 dark:border-slate-700/50">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg flex items-center justify-center">
              <Bot className="w-5 h-5 text-white" />
            </div>
            <span className="text-xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
              NodePath
            </span>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setSidebarOpen(false)}
            className="lg:hidden"
          >
            <X className="w-4 h-4" />
          </Button>
        </div>
        
        <nav className="mt-8 px-4 space-y-2">
          {filteredNavigation.map((item) => {
            const Icon = item.icon;
            return (
              <a
                key={item.name}
                href={item.href}
                className={cn(
                  "group flex items-center px-3 py-2.5 text-sm font-medium rounded-lg transition-all duration-200",
                  item.current
                    ? "bg-gradient-to-r from-blue-500/10 to-purple-500/10 text-blue-700 dark:text-blue-300 border border-blue-200/50 dark:border-blue-700/50"
                    : "text-slate-600 dark:text-slate-300 hover:bg-slate-100/50 dark:hover:bg-slate-800/50 hover:text-slate-900 dark:hover:text-white"
                )}
              >
                <Icon className={cn(
                  "mr-3 h-5 w-5 transition-colors",
                  item.current ? "text-blue-600 dark:text-blue-400" : "text-slate-400 group-hover:text-slate-600 dark:group-hover:text-slate-300"
                )} />
                {item.name}
                {item.current && (
                  <div className="ml-auto w-2 h-2 bg-blue-500 rounded-full animate-pulse" />
                )}
              </a>
            );
          })}
        </nav>
        

      </div>

      {/* Mobile sidebar overlay */}
      {sidebarOpen && (
        <div 
          className="fixed inset-0 z-40 bg-black/20 backdrop-blur-sm lg:hidden" 
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Main content */}
      <div className={cn(
        "transition-all duration-300 ease-in-out",
        sidebarOpen ? "lg:ml-64" : "ml-0"
      )}>
        {/* Top bar */}
        <TopBar 
          sidebarOpen={sidebarOpen} 
          setSidebarOpen={setSidebarOpen} 
        />

        {/* Page content */}
        <main className="p-6">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/device-settings" element={<DeviceSettings />} />
            <Route path="/profile" element={<Profile />} />
            <Route path="/billings" element={<Billings />} />
            <Route path="/set-stage" element={<SetStage />} />
            <Route path="/flow-builder" element={
              <DeviceRequiredWrapper>
                <FlowBuilder />
              </DeviceRequiredWrapper>
            } />
            <Route path="/flow-manager" element={
              <DeviceRequiredWrapper>
                <FlowManager />
              </DeviceRequiredWrapper>
            } />
            <Route path="/chatbot-ai" element={
              <DeviceRequiredWrapper>
                <Analytics />
              </DeviceRequiredWrapper>
            } />
            <Route path="/whatsapp-bot" element={
              <DeviceRequiredWrapper>
                <WhatsAppBot />
              </DeviceRequiredWrapper>
            } />
          </Routes>
        </main>
      </div>
    </div>
  );
}

const App = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <AuthProvider>
          <DeviceProvider>
            <Router>
              <Routes>
                {/* Public routes */}
                <Route path="/login" element={<Login />} />
                <Route path="/register" element={<Register />} />
                
                {/* Protected routes */}
                <Route path="/*" element={
                  <ProtectedRoute>
                    <AppContent />
                  </ProtectedRoute>
                } />
              </Routes>
              
              <Toaster />
              <Sonner />
            </Router>
          </DeviceProvider>
        </AuthProvider>
      </TooltipProvider>
    </QueryClientProvider>
  );
};

export default App;