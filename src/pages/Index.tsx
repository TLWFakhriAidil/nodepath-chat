import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Bot, Workflow, MessageSquare } from 'lucide-react';

import ChatbotBuilder from '@/components/ChatbotBuilder';
import ChatSimulation from '@/components/ChatSimulation';

const Index = () => {
  return (
    <div className="min-h-screen bg-background">
      <Tabs defaultValue="builder" className="h-screen flex flex-col">
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
          </div>
        </div>

        {/* Content */}
        <div className="flex-1">
          <TabsContent value="builder" className="h-full m-0">
            <ChatbotBuilder />
          </TabsContent>
          
          <TabsContent value="simulation" className="h-full m-0 p-6">
            <div className="flex items-center justify-center h-full">
              <div className="text-center space-y-6">
                <h2 className="text-2xl font-bold text-foreground">Test Your Chatbot</h2>
                <p className="text-muted-foreground max-w-md">
                  This is a simulation of how your chatbot will interact with users. 
                  In a real implementation, responses would be generated from your flow logic.
                </p>
                <ChatSimulation />
              </div>
            </div>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
};

export default Index;
