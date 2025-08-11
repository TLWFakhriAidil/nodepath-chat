import { useState } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from '@/components/ui/button';
import { Bot, Workflow, MessageSquare, Upload } from 'lucide-react';
import ChatbotBuilder from '@/components/ChatbotBuilder';
import ChatSimulation from '@/components/ChatSimulation';

const Index = () => {
  const [activeTab, setActiveTab] = useState('builder');
  const [testFlowId, setTestFlowId] = useState<string | null>(null);

  const handleTestFlow = (flowId: string) => {
    setTestFlowId(flowId);
    setActiveTab('simulation');
  };

  return (
    <div className="min-h-screen bg-background">
      <Tabs value={activeTab} onValueChange={setActiveTab} className="h-screen flex flex-col">
        {/* Header */}
        <div className="bg-card border-b border-border px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <div className="bg-gradient-primary p-2 rounded-lg">
                <Bot className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-foreground">ChatBot Builder</h1>
                <p className="text-sm text-muted-foreground">Build intelligent conversational flows</p>
              </div>
            </div>
            
            <div className="flex items-center gap-4">
              <TabsList className="grid w-[400px] grid-cols-2">
                <TabsTrigger value="builder" className="flex items-center space-x-2">
                  <Workflow className="w-4 h-4" />
                  <span>Flow Builder</span>
                </TabsTrigger>
                <TabsTrigger value="simulation" className="flex items-center space-x-2">
                  <MessageSquare className="w-4 h-4" />
                  <span>Test Chat</span>
                </TabsTrigger>
              </TabsList>
              
              <Button variant="outline" className="flex items-center gap-2">
                <Upload className="w-4 h-4" />
                Media Manager
              </Button>
            </div>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1">
          <TabsContent value="builder" className="h-full m-0">
            <ChatbotBuilder onTestFlow={handleTestFlow} />
          </TabsContent>
          
          <TabsContent value="simulation" className="h-full m-0">
            <ChatSimulation key={testFlowId} preselectedFlowId={testFlowId} />
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
};

export default Index;
