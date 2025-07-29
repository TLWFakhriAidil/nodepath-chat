import { useState } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { AppSidebar } from '@/components/AppSidebar';
import ChatbotBuilder from '@/components/ChatbotBuilder';
import ChatSimulation from '@/components/ChatSimulation';
import FlowManager from '@/components/FlowManager';
import MediaManager from './MediaManager';

const Index = () => {
  const [testFlowId, setTestFlowId] = useState<string | null>(null);

  const handleTestFlow = (flowId: string) => {
    setTestFlowId(flowId);
    // Navigate to test page
    window.location.hash = '#/test';
  };

  const handleCreateNewFlow = () => {
    // Navigate to builder
    window.location.hash = '#/';
  };

  return (
    <Router>
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
                <Route 
                  path="/media" 
                  element={<MediaManager />} 
                />
              </Routes>
            </main>
          </div>
        </div>
      </SidebarProvider>
    </Router>
  );
};

export default Index;
