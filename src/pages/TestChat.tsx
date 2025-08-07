import React, { useState, useEffect, useRef } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { MessageCircle, Send, Trash2, Sparkles } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import { getFlows } from '@/lib/mysqlStorage';

interface FlowNode {
  id: string;
  type: string;
  data: {
    label?: string;
    instance?: string;
    openRouterKey?: string;
    node_type?: string;
  };
}

interface ChatMessage {
  role: 'USER' | 'BOT';
  content: string;
  timestamp: string;
}

interface ConversationData {
  id: string;
  instance: string;
  open_router_key: string;
  conv_last: any[];
  conv_current: string;
}

export default function TestChat() {
  const [flows, setFlows] = useState<any[]>([]);
  const [aiNodes, setAiNodes] = useState<FlowNode[]>([]);
  const [selectedNodeId, setSelectedNodeId] = useState<string>('');
  const [selectedNode, setSelectedNode] = useState<FlowNode | null>(null);
  const [userMessage, setUserMessage] = useState('');
  const [conversation, setConversation] = useState<ConversationData | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const { toast } = useToast();

  useEffect(() => {
    loadFlows();
  }, []);

  useEffect(() => {
    if (selectedNodeId) {
      const node = aiNodes.find(n => n.id === selectedNodeId);
      setSelectedNode(node || null);
      if (node) {
        loadOrCreateConversation(node);
      }
    }
  }, [selectedNodeId, aiNodes]);

  useEffect(() => {
    scrollToBottom();
  }, [conversation]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const loadFlows = async () => {
    try {
      const data = await getFlows();
      setFlows(data || []);

      // Extract AI prompt nodes from all flows
      const allAiNodes: FlowNode[] = [];
      (data || []).forEach(flow => {
        const nodes = Array.isArray(flow.nodes) ? flow.nodes : [];
        const promptNodes = nodes.filter((node: any) => 
          node.type === 'prompt' || node.data?.node_type === 'ai_prompt' || node.type === 'manual'
        );
        allAiNodes.push(...promptNodes.map((node: any) => ({
          ...node,
          flowName: flow.name
        })));
      });

      setAiNodes(allAiNodes);
    } catch (error) {
      console.error('Error loading flows:', error);
      toast({
        title: "Error",
        description: "Failed to load flows",
        variant: "destructive"
      });
    }
  };

  const loadOrCreateConversation = async (node: FlowNode) => {
    try {
      // Since MySQL is disconnected, use localStorage
      const conversations = JSON.parse(localStorage.getItem('test_conversations') || '{}');
      
      if (conversations[node.id]) {
        setConversation(conversations[node.id]);
      } else {
        // Create new conversation with data from the flow node
        const newConversation = {
          id: node.id,
          instance: node.data.instance || 'default',
          open_router_key: node.data.openRouterKey || '',
          conv_last: [],
          conv_current: ''
        };

        conversations[node.id] = newConversation;
        localStorage.setItem('test_conversations', JSON.stringify(conversations));
        setConversation(newConversation);
      }
    } catch (error) {
      console.error('Error loading conversation:', error);
      toast({
        title: "Error",
        description: "Failed to load conversation data",
        variant: "destructive"
      });
    }
  };

  const sendMessage = async () => {
    if (!userMessage.trim() || !selectedNode || !conversation) return;

    setIsLoading(true);

    try {
      // Add user message to conversation
      const userMsg: ChatMessage = {
        role: 'USER',
        content: userMessage.trim(),
        timestamp: new Date().toISOString()
      };

      let botReply = '';

      if (selectedNode.type === 'manual') {
        // Handle manual node - use predefined response
        botReply = selectedNode.data.label || 'Manual response configured in Flow Builder';
      } else if (selectedNode.type === 'prompt') {
        // AI functionality has been removed
        botReply = `AI functionality has been removed. This was an AI prompt node with instance: ${conversation.instance}`;
      } else {
        // Default fallback for other node types
        botReply = `Node type "${selectedNode.type}" response: ${selectedNode.data.label || 'Default response'}`;
      }

      const botMsg: ChatMessage = {
        role: 'BOT',
        content: botReply,
        timestamp: new Date().toISOString()
      };

      // Update conversation in localStorage
      const updatedConvLast = [...(conversation.conv_last || []), userMsg, botMsg];
      const updatedConversation = {
        ...conversation,
        conv_last: updatedConvLast,
        conv_current: botReply
      };
      
      const conversations = JSON.parse(localStorage.getItem('test_conversations') || '{}');
      conversations[conversation.id] = updatedConversation;
      localStorage.setItem('test_conversations', JSON.stringify(conversations));

      // Update local state
      setConversation(updatedConversation);
      setUserMessage('');

      toast({
        title: "Message sent",
        description: selectedNode.type === 'manual' ? "Manual response delivered" : "Mock response generated"
      });

    } catch (error) {
      console.error('Error sending message:', error);
      toast({
        title: "Error",
        description: error instanceof Error ? error.message : "Failed to send message",
        variant: "destructive"
      });
    } finally {
      setIsLoading(false);
    }
  };

  const clearConversation = async () => {
    if (!conversation) return;

    try {
      const conversations = JSON.parse(localStorage.getItem('test_conversations') || '{}');
      conversations[conversation.id] = {
        ...conversation,
        conv_last: [],
        conv_current: ''
      };
      localStorage.setItem('test_conversations', JSON.stringify(conversations));

      setConversation({
        ...conversation,
        conv_last: [],
        conv_current: ''
      });

      toast({
        title: "Conversation cleared",
        description: "Chat history has been reset"
      });

    } catch (error) {
      console.error('Error clearing conversation:', error);
      toast({
        title: "Error",
        description: "Failed to clear conversation",
        variant: "destructive"
      });
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  return (
    <div className="container mx-auto p-6 max-w-4xl">
      <div className="mb-6">
        <h1 className="text-3xl font-bold mb-2 flex items-center gap-2">
          <MessageCircle className="w-8 h-8 text-primary" />
          Test Chat
        </h1>
        <p className="text-muted-foreground">
          Test chatbot nodes with simulated conversations. AI functionality has been removed.
        </p>
      </div>

      <Card className="h-[600px] flex flex-col">
        <CardHeader className="pb-4">
          <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Sparkles className="w-5 h-5" />
              Chat Simulation (No AI)
            </CardTitle>
            <div className="flex gap-2">
              <Select value={selectedNodeId} onValueChange={setSelectedNodeId}>
                <SelectTrigger className="w-[300px]">
                  <SelectValue placeholder="Select a prompt node to test" />
                </SelectTrigger>
                <SelectContent>
                  {aiNodes.map((node) => (
                    <SelectItem key={node.id} value={node.id}>
                      {(node as any).flowName} - {node.data.label || `Node ${node.id.slice(0, 8)}`}
                      {node.type === 'manual' && ' (Manual)'}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {conversation && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={clearConversation}
                  className="flex items-center gap-2"
                >
                  <Trash2 className="w-4 h-4" />
                  Clear
                </Button>
              )}
            </div>
          </div>

          {selectedNode && (
            <div className="mt-4 p-3 bg-muted rounded-lg text-sm">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div>
                  <span className="font-medium">Instance:</span>
                  <p className="text-muted-foreground mt-1">
                    {conversation?.instance || selectedNode.data.instance || 'default'}
                  </p>
                </div>
                <div>
                  <span className="font-medium">Status:</span>
                  <p className="text-muted-foreground mt-1">
                    AI Disabled - Manual responses only
                  </p>
                </div>
              </div>
            </div>
          )}
        </CardHeader>

        <CardContent className="flex-1 flex flex-col">
          {/* Chat Messages */}
          <div className="flex-1 overflow-y-auto mb-4 p-4 bg-muted/20 rounded-lg">
            {!selectedNode ? (
              <div className="text-center text-muted-foreground py-8">
                Select a prompt node from the dropdown above to start testing
              </div>
            ) : !conversation?.conv_last?.length ? (
              <div className="text-center text-muted-foreground py-8">
                No messages yet. Type a message below to start the conversation.
              </div>
            ) : (
              <div className="space-y-4">
                {conversation.conv_last.map((message, index) => (
                  <div
                    key={index}
                    className={`flex ${message.role === 'USER' ? 'justify-end' : 'justify-start'}`}
                  >
                    <div
                      className={`max-w-[80%] p-3 rounded-lg ${
                        message.role === 'USER'
                          ? 'bg-primary text-primary-foreground'
                          : 'bg-card border'
                      }`}
                    >
                      <div className="text-sm font-medium mb-1">
                        {message.role === 'USER' ? 'User' : 'Bot'}
                      </div>
                      <div className="whitespace-pre-wrap">{message.content}</div>
                      <div className="text-xs opacity-70 mt-2">
                        {new Date(message.timestamp).toLocaleTimeString()}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          {/* Message Input */}
          {selectedNode && (
            <div className="flex gap-3">
              <Textarea
                value={userMessage}
                onChange={(e) => setUserMessage(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="Type your test message here..."
                className="flex-1 min-h-[80px] resize-none"
                disabled={isLoading}
              />
              <Button
                onClick={sendMessage}
                disabled={!userMessage.trim() || isLoading}
                size="lg"
                className="px-6"
              >
                <Send className="w-4 h-4" />
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}