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
      <AppSidebar />
              
              {/* Main content area with responsive left margin for sidebar */}
              <div className={`${isCollapsed ? 'ml-14' : 'ml-64'} min-h-screen flex flex-col transition-all duration-300`}>
                {/* Header with sidebar trigger */}
                <header className="h-12 flex items-center border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
                  <SidebarTrigger className="ml-4" />
                </header>

                {/* Main content */}
                <main className="flex-1 overflow-hidden">
                  <Routes>
                    <Route 
                      path="/" 
                      element={<ChatbotBuilder onTestFlow={handleTestFlow} />} 
                    />
                    <Route 
                      path="/flows" 
                      element={
                        <FlowManager 
                          onCreateNew={handleCreateNewFlow}
                          onTestFlow={handleTestFlow}
                        />
                      } 
                    />
                    <Route 
                      path="/test" 
                      element={<ChatSimulation key={testFlowId} preselectedFlowId={testFlowId} />} 
                    />
                    <Route path="/test-chat" element={
                      <Suspense fallback={<div className="flex items-center justify-center h-screen">Loading...</div>}>
                        <TestChat />
                      </Suspense>
                    } />
                    <Route path="/media" element={
                      <Suspense fallback={<div className="flex items-center justify-center h-screen">Loading...</div>}>
                        <MediaManager />
                      </Suspense>
                    } />
                     <Route path="/analytics" element={
                       <Suspense fallback={<div className="flex items-center justify-center h-screen">Loading...</div>}>
                         <LeadAnalytics />
                       </Suspense>
                     } />
                     {/* ADD ALL CUSTOM ROUTES ABOVE THE CATCH-ALL "*" ROUTE */}
                    <Route path="*" element={<NotFound />} />
                  </Routes>
                </main>
              </div>
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
