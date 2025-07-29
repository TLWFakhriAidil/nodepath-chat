import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { lazy, Suspense, useState } from "react";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { AppSidebar } from '@/components/AppSidebar';
import Index from "./pages/Index";
import NotFound from "./pages/NotFound";
import ChatbotBuilder from '@/components/ChatbotBuilder';
import ChatSimulation from '@/components/ChatSimulation';
import FlowManager from '@/components/FlowManager';

// Lazy load MediaManager to avoid importing supabase on initial load
const MediaManager = lazy(() => import("./pages/MediaManager"));

const queryClient = new QueryClient();

const App = () => {
  const [testFlowId, setTestFlowId] = useState<string | null>(null);

  const handleTestFlow = (flowId: string) => {
    setTestFlowId(flowId);
  };

  const handleCreateNewFlow = () => {
    // Will be handled by navigation
  };

  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <Toaster />
        <Sonner />
        <BrowserRouter>
          <SidebarProvider>
            <div className="min-h-screen flex w-full bg-background">
              <AppSidebar />
              
              <div className="flex-1 flex flex-col">
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
                    <Route path="/media" element={
                      <Suspense fallback={<div className="flex items-center justify-center h-screen">Loading...</div>}>
                        <MediaManager />
                      </Suspense>
                    } />
                    {/* ADD ALL CUSTOM ROUTES ABOVE THE CATCH-ALL "*" ROUTE */}
                    <Route path="*" element={<NotFound />} />
                  </Routes>
                </main>
              </div>
            </div>
          </SidebarProvider>
        </BrowserRouter>
      </TooltipProvider>
    </QueryClientProvider>
  );
};

export default App;
