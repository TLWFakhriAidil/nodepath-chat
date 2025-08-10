import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { lazy, Suspense, useState } from "react";
import { SidebarProvider, SidebarTrigger, useSidebar } from "@/components/ui/sidebar";
import { AppSidebar } from '@/components/AppSidebar';
import Index from "./pages/Index";
import NotFound from "./pages/NotFound";
import ChatbotBuilder from '@/components/ChatbotBuilder';
import ChatSimulation from '@/components/ChatSimulation';
import FlowManager from '@/components/FlowManager';

// Lazy load pages to avoid importing supabase on initial load
const MediaManager = lazy(() => import("./pages/MediaManager"));
const LeadAnalytics = lazy(() => import("./pages/LeadAnalytics"));

const TestChat = lazy(() => import("./pages/TestChat"));

const queryClient = new QueryClient();

const AppContent = () => {
  const [testFlowId, setTestFlowId] = useState<string | null>(null);
  const { state } = useSidebar();
  const isCollapsed = state === 'collapsed';

  const handleTestFlow = (flowId: string) => {
    setTestFlowId(flowId);
  };

  const handleCreateNewFlow = () => {
    // Will be handled by navigation
  };

  return (
    <div className="min-h-screen bg-background">
      <Routes>
        {/* Flow Builder with navigation sidebar */}
        <Route 
          path="/" 
          element={
            <div className="min-h-screen bg-background flex">
              <AppSidebar />
              <div className="flex-1 min-h-screen ml-56">
                <ChatbotBuilder onTestFlow={handleTestFlow} />
              </div>
            </div>
          } 
        />
        
        {/* Other routes use the sidebar layout */}
        <Route path="/flows" element={
          <div className="min-h-screen bg-background flex">
            <AppSidebar />
            
            {/* Main content area */}
            <div className="flex-1 min-h-screen flex flex-col">
              {/* Header with sidebar trigger */}
              <header className="h-12 flex items-center border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-4">
                <SidebarTrigger />
              </header>

              {/* Main content */}
              <main className="flex-1 overflow-auto">
                <FlowManager 
                  onCreateNew={handleCreateNewFlow}
                  onTestFlow={handleTestFlow}
                />
              </main>
            </div>
          </div>
        } />
        
        <Route path="/test" element={
          <div className="min-h-screen bg-background flex">
            <AppSidebar />
            <div className="flex-1 min-h-screen flex flex-col">
              <header className="h-12 flex items-center border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-4">
                <SidebarTrigger />
              </header>
              <main className="flex-1 overflow-auto">
                <ChatSimulation key={testFlowId} preselectedFlowId={testFlowId} />
              </main>
            </div>
          </div>
        } />
        
        <Route path="/test-chat" element={
          <div className="min-h-screen bg-background flex">
            <AppSidebar />
            <div className="flex-1 min-h-screen flex flex-col">
              <header className="h-12 flex items-center border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-4">
                <SidebarTrigger />
              </header>
              <main className="flex-1 overflow-auto">
                <Suspense fallback={<div className="flex items-center justify-center h-screen">Loading...</div>}>
                  <TestChat />
                </Suspense>
              </main>
            </div>
          </div>
        } />
        
        <Route path="/media" element={
          <div className="min-h-screen bg-background flex">
            <AppSidebar />
            <div className="flex-1 min-h-screen flex flex-col">
              <header className="h-12 flex items-center border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-4">
                <SidebarTrigger />
              </header>
              <main className="flex-1 overflow-auto">
                <Suspense fallback={<div className="flex items-center justify-center h-screen">Loading...</div>}>
                  <MediaManager />
                </Suspense>
              </main>
            </div>
          </div>
        } />
        
        <Route path="/analytics" element={
          <div className="min-h-screen bg-background flex">
            <AppSidebar />
            <div className="flex-1 min-h-screen flex flex-col">
              <header className="h-12 flex items-center border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-4">
                <SidebarTrigger />
              </header>
              <main className="flex-1 overflow-auto">
                <Suspense fallback={<div className="flex items-center justify-center h-screen">Loading...</div>}>
                  <LeadAnalytics />
                </Suspense>
              </main>
            </div>
          </div>
        } />
        
        <Route path="*" element={<NotFound />} />
      </Routes>
    </div>
  );
};

const App = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <Toaster />
        <Sonner />
        <BrowserRouter>
          <SidebarProvider>
            <AppContent />
          </SidebarProvider>
        </BrowserRouter>
      </TooltipProvider>
    </QueryClientProvider>
  );
};

export default App;
