import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { lazy, Suspense, useState } from "react";
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



const App = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <TooltipProvider>
        <Toaster />
        <Sonner />
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<Index />} />
            <Route path="/flow-builder" element={<ChatbotBuilder />} />
            <Route path="/chat-simulation" element={<ChatSimulation />} />
            <Route path="/flow-manager" element={<FlowManager />} />
            <Route path="/flows" element={<FlowManager />} />
            <Route path="/test" element={<ChatSimulation />} />
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
            <Route path="*" element={<NotFound />} />
          </Routes>
        </BrowserRouter>
      </TooltipProvider>
    </QueryClientProvider>
  );
};

export default App;
