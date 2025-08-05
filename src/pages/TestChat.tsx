import React, { useState, useEffect, useRef } from 'react';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { MessageCircle, Send, Trash2, Sparkles } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import { supabase } from '@/integrations/supabase/client';

interface FlowNode {
  id: string;
  type: string;
  data: {
    label?: string;
    systemPrompt?: string;
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
  system_prompt: string;
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
      const { data, error } = await supabase
        .from('chatbot_flows')
        .select('*')
        .order('updated_at', { ascending: false });

      if (error) throw error;

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
      const { data, error } = await supabase
        .from('chatbot_executions_nodepath')
        .select('*')
        .eq('id', node.id)
        .maybeSingle();

      if (error && error.code !== 'PGRST116') throw error;

      if (data) {
        setConversation({
          ...data,
          conv_last: Array.isArray(data.conv_last) ? data.conv_last : []
        });
      } else {
        // Create new conversation record
        const newConversation = {
          id: node.id,
          system_prompt: node.data.systemPrompt || '',
          instance: node.data.instance || '',
          open_router_key: node.data.openRouterKey || '',
          conv_last: [],
          conv_current: ''
        };

        const { error: insertError } = await supabase
          .from('chatbot_executions_nodepath')
          .insert(newConversation);

        if (insertError) throw insertError;

        setConversation(newConversation);
      }
    } catch (error) {
      console.error('Error loading conversation:', error);
      toast({
        title: "Error",
        description: "Failed to load conversation",
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
      } else {
        // Handle AI prompt node - call AI with conversation history
        const { data, error } = await supabase.functions.invoke('test-ai-chat', {
          body: {
            systemPrompt: conversation.system_prompt,
            userMessage: userMessage.trim(),
            instance: conversation.instance,
            openRouterKey: conversation.open_router_key,
            conversationHistory: conversation.conv_last || []
          }
        });

        if (error) throw error;
        if (!data.success) throw new Error(data.error);

        botReply = data.aiReply;
      }

      const botMsg: ChatMessage = {
        role: 'BOT',
        content: botReply,
        timestamp: new Date().toISOString()
      };

      // Update conversation in database
      const updatedConvLast = [...(conversation.conv_last || []), userMsg, botMsg];
      
      const { error: updateError } = await supabase
        .from('chatbot_executions_nodepath')
        .update({
          conv_last: updatedConvLast,
          conv_current: botReply,
          updated_at: new Date().toISOString()
        })
        .eq('id', conversation.id);

      if (updateError) throw updateError;

      // Update local state
      setConversation({
        ...conversation,
        conv_last: updatedConvLast,
        conv_current: botReply
      });

      setUserMessage('');

      toast({
        title: "Message sent",
        description: selectedNode.type === 'manual' ? "Manual response delivered" : "AI response generated"
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
      const { error } = await supabase
        .from('chatbot_executions_nodepath')
        .update({
          conv_last: [],
          conv_current: '',
          updated_at: new Date().toISOString()
        })
        .eq('id', conversation.id);

      if (error) throw error;

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
          Test AI prompt nodes with simulated conversations. No real WhatsApp messages are sent.
        </p>
      </div>

      <Card className="h-[600px] flex flex-col">
        <CardHeader className="pb-4">
          <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <Sparkles className="w-5 h-5" />
              AI Chat Simulation
            </CardTitle>
            <div className="flex gap-2">
              <Select value={selectedNodeId} onValueChange={setSelectedNodeId}>
                <SelectTrigger className="w-[300px]">
                  <SelectValue placeholder="Select an AI prompt node to test" />
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
              <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                <div>
                  <span className="font-medium">System Prompt:</span>
                  <p className="text-muted-foreground mt-1">
                    {selectedNode.data.systemPrompt || 'Default assistant prompt'}
                  </p>
                </div>
                <div>
                  <span className="font-medium">Instance:</span>
                  <p className="text-muted-foreground mt-1">
                    {selectedNode.data.instance || 'default'}
                  </p>
                </div>
                <div>
                  <span className="font-medium">Router Key:</span>
                  <p className="text-muted-foreground mt-1">
                    {selectedNode.data.openRouterKey || 'none'}
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
                Select an AI prompt node from the dropdown above to start testing
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