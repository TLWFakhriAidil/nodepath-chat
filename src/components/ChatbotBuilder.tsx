import React, { useCallback, useState, useEffect } from 'react';
import {
  ReactFlow,
  addEdge,
  MiniMap,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  Connection,
  Edge,
  Node,
  NodeTypes,
  MarkerType,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { MessageSquare, GitBranch, Clock, Play, Download, Upload, Image, Mic, Video, Save, Sparkles, X, Send, Bot } from 'lucide-react';
import { ChatbotFlow } from '@/types/chatbot';
import { saveFlow, getFlows, getFlow } from '@/lib/localStorage';
import { useToast } from '@/hooks/use-toast';

import MessageNode from './nodes/MessageNode';
import ConditionNode from './nodes/ConditionNode';
import DelayNode from './nodes/DelayNode';
import StartNode from './nodes/StartNode';
import ImageNode from './nodes/ImageNode';
import AudioNode from './nodes/AudioNode';
import VideoNode from './nodes/VideoNode';
import ManualNode from './nodes/ManualNode';
import PromptNode from './nodes/PromptNode';
import StageNode from './nodes/StageNode';

const nodeTypes: NodeTypes = {
  message: MessageNode,
  condition: ConditionNode,
  delay: DelayNode,
  start: StartNode,
  image: ImageNode,
  audio: AudioNode,
  video: VideoNode,
  manual: ManualNode,
  prompt: PromptNode,
  stage: StageNode,
};

const initialNodes: Node[] = [
  {
    id: 'start-1',
    type: 'start',
    position: { x: 250, y: 100 },
    data: { label: 'Start' },
  },
];

const initialEdges: Edge[] = [];

export default function ChatbotBuilder({ onTestFlow, flowId }: { onTestFlow?: (flowId: string) => void; flowId?: string | null }) {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [selectedNodeType, setSelectedNodeType] = useState<string | null>(null);
  const [flowName, setFlowName] = useState('');
  const [currentFlowId, setCurrentFlowId] = useState<string | null>(null);
  const [fieldInstance, setFieldInstance] = useState('');
  const [openRouterApiKey, setOpenRouterApiKey] = useState('');
  const { toast } = useToast();

  const deleteNode = useCallback(
    (nodeId: string) => {
      setNodes((nds) => nds.filter((node) => node.id !== nodeId));
      setEdges((eds) => eds.filter((edge) => edge.source !== nodeId && edge.target !== nodeId));
    },
    [setNodes, setEdges]
  );

  const updateNodeData = useCallback(
    (nodeId: string, newData: any) => {
      setNodes((nds) =>
        nds.map((node) =>
          node.id === nodeId
            ? { ...node, data: { ...node.data, ...newData } }
            : node
        )
      );
    },
    [setNodes]
  );

  const onConnect = useCallback(
    (params: Edge | Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  );

  const onEdgesDelete = useCallback(
    (edgesToDelete: Edge[]) => {
      setEdges((eds) => eds.filter((edge) => !edgesToDelete.find((e) => e.id === edge.id)));
    },
    [setEdges]
  );

  const addNode = useCallback(
    (type: string) => {
      const newNode: Node = {
        id: `${type}-${Date.now()}`,
        type,
        position: { x: Math.random() * 400 + 100, y: Math.random() * 400 + 200 },
        data: {
          label: type === 'message' ? 'New Message' : 
                 type === 'condition' ? 'New Condition' : 
                 type === 'delay' ? 'New Delay' :
                 type === 'image' ? 'New Image' :
                 type === 'audio' ? 'New Audio' :
                 type === 'video' ? 'New Video' :
                 type === 'manual' ? 'Manual Response' :
                 type === 'prompt' ? 'AI Prompt' :
                 'New Node',
          message: type === 'message' ? 'Enter your message here...' : undefined,
          condition: type === 'condition' ? 'user_input contains "yes"' : undefined,
          delay: type === 'delay' ? 5 : undefined,
          imageUrl: type === 'image' ? '' : undefined,
          caption: type === 'image' || type === 'video' ? 'Caption...' : undefined,
          audioUrl: type === 'audio' ? '' : undefined,
          videoUrl: type === 'video' ? '' : undefined,
          duration: type === 'audio' || type === 'video' ? (type === 'audio' ? 30 : 60) : undefined,
          expectedInput: type === 'manual' ? '' : undefined,
          responseOutput: type === 'manual' ? '' : undefined,
          systemPrompt: type === 'prompt' ? 'You are a helpful assistant that responds clearly and concisely.' : undefined,
          stageName: type === 'stage' ? '' : undefined,
          node_type: type === 'prompt' ? 'ai_prompt' : undefined,
          onDelete: deleteNode,
          onUpdate: updateNodeData,
        },
      };
      setNodes((nds) => nds.concat(newNode));
    },
    [setNodes, deleteNode, updateNodeData]
  );

  // Load flow data when flowId is provided
  useEffect(() => {
    const loadFlowData = async () => {
      if (!flowId) return;
      
      try {
        const flowData = await getFlow(flowId);
        if (flowData) {
          setFlowName(flowData.name);
          setCurrentFlowId(flowData.id);
          setFieldInstance(flowData.globalInstance || '');
          setOpenRouterApiKey(flowData.globalOpenRouterKey || '');
          
          // Load nodes with proper callbacks
          const loadedNodes = flowData.nodes.map((node: any) => ({
            ...node,
            data: {
              ...node.data,
              onDelete: deleteNode,
              onUpdate: updateNodeData,
            }
          }));
          
          setNodes(loadedNodes);
          setEdges(flowData.edges || []);
          
          toast({
            title: "Flow loaded",
            description: `"${flowData.name}" has been loaded for editing`,
            variant: "default"
          });
        }
      } catch (error) {
        console.error('Error loading flow:', error);
        toast({
          title: "Load failed",
          description: "Failed to load flow data",
          variant: "destructive"
        });
      }
    };
    
    loadFlowData();
  }, [flowId, setNodes, setEdges, toast, deleteNode, updateNodeData]);

  const saveFlowToStorage = useCallback(async () => {
    if (!flowName.trim()) {
      toast({
        title: "Flow name required",
        description: "Please enter a name for your flow",
        variant: "destructive"
      });
      return;
    }

    // Check if we have any nodes besides the default start node
    if (nodes.length <= 1) {
      toast({
        title: "No flow to save",
        description: "Please add some nodes and connections to your flow before saving",
        variant: "destructive"
      });
      return;
    }

    const flowData: ChatbotFlow = {
      id: currentFlowId || `flow_${Date.now()}_${Math.random().toString(36).substring(2)}`,
      name: flowName,
      description: `Chatbot flow: ${flowName}`,
      globalInstance: fieldInstance,
      globalOpenRouterKey: openRouterApiKey,
      nodes: nodes.map(node => ({
        id: node.id,
        type: node.type as any,
        position: node.position,
        data: node.data
      })),
      edges: edges.map(edge => ({
        id: edge.id || `${edge.source}-${edge.target}`,
        source: edge.source,
        target: edge.target,
        sourceHandle: edge.sourceHandle,
        targetHandle: edge.targetHandle
      })),
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };
    
    try {
      await saveFlow(flowData);
      setCurrentFlowId(flowData.id);
      
      toast({
        title: "Flow saved",
        description: `"${flowName}" has been saved successfully`,
        variant: "default"
      });
    } catch (error) {
      console.error('Error saving flow:', error);
      
      toast({
        title: "Save failed",
        description: error instanceof Error ? error.message : 'Failed to save flow to database. Check console for details.',
        variant: "destructive"
      });
      
      // Still set the current flow ID for local editing
      setCurrentFlowId(flowData.id);
    }
  }, [flowName, currentFlowId, nodes, edges, toast]);

  const exportFlow = useCallback(() => {
    const flowData = {
      nodes,
      edges,
      timestamp: new Date().toISOString(),
    };
    
    const dataStr = JSON.stringify(flowData, null, 2);
    const dataBlob = new Blob([dataStr], { type: 'application/json' });
    const url = URL.createObjectURL(dataBlob);
    
    const link = document.createElement('a');
    link.href = url;
    link.download = `chatbot-flow-${Date.now()}.json`;
    link.click();
    
    URL.revokeObjectURL(url);
  }, [nodes, edges]);

  const importFlow = useCallback(() => {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.json';
    input.onchange = (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (!file) return;
      
      const reader = new FileReader();
      reader.onload = (event) => {
        try {
          const flowData = JSON.parse(event.target?.result as string);
          
          if (flowData.nodes && Array.isArray(flowData.nodes)) {
            // Clear current flow
            setNodes([]);
            setEdges([]);
            
            // Import nodes
            const importedNodes = flowData.nodes.map((node: any) => ({
              id: node.id || `node_${Date.now()}_${Math.random().toString(36).substring(2)}`,
              type: node.type || 'message',
              position: node.position || { x: Math.random() * 400, y: Math.random() * 400 },
              data: {
                ...node.data,
                onDelete: deleteNode,
                onUpdate: updateNodeData,
              }
            }));
            
            // Import edges if they exist
            const importedEdges = (flowData.edges || []).map((edge: any) => ({
              id: edge.id || `${edge.source}-${edge.target}`,
              source: edge.source,
              target: edge.target,
              sourceHandle: edge.sourceHandle,
              targetHandle: edge.targetHandle
            }));
            
            setNodes(importedNodes);
            setEdges(importedEdges);
            
            toast({
              title: "Flow imported",
              description: `Successfully imported ${importedNodes.length} nodes and ${importedEdges.length} connections`,
              variant: "default"
            });
          } else {
            throw new Error('Invalid flow format');
          }
        } catch (error) {
          console.error('Error importing flow:', error);
          toast({
            title: "Import failed",
            description: "Failed to import flow. Please check the file format.",
            variant: "destructive"
          });
        }
      };
      reader.readAsText(file);
    };
    input.click();
  }, [setNodes, setEdges, deleteNode, updateNodeData, toast]);

  const [showTestModal, setShowTestModal] = useState(false);
  const [testMessages, setTestMessages] = useState<Array<{type: 'user' | 'bot', content: string}>>([]);
  const [testInput, setTestInput] = useState('');

  const sendTestMessage = useCallback(() => {
    if (!testInput.trim()) return;
    
    setTestMessages(prev => [...prev, { type: 'user', content: testInput }]);
    
    // Simple echo response for demo
    setTimeout(() => {
      setTestMessages(prev => [...prev, { type: 'bot', content: `Echo: ${testInput}` }]);
    }, 500);
    
    setTestInput('');
  }, [testInput]);

  const testFlow = useCallback(() => {
    if (!currentFlowId) {
      saveFlowToStorage();
      return;
    }
    setShowTestModal(true);
    setTestMessages([{ type: 'bot', content: 'Test flow started. Type a message to begin.' }]);
  }, [currentFlowId, saveFlowToStorage]);

  const nodeTypeButtons = [
    { type: 'message', label: 'Send Message', icon: MessageSquare, color: 'bg-gradient-to-r from-blue-500 to-cyan-500', gradient: 'gradient-primary' },
    { type: 'manual', label: 'Manual Response', icon: MessageSquare, color: 'bg-gradient-to-r from-pink-500 to-rose-500', gradient: 'gradient-pink' },
    { type: 'prompt', label: 'AI Prompt', icon: Sparkles, color: 'bg-gradient-to-r from-purple-500 to-violet-500', gradient: 'gradient-purple' },
    { type: 'stage', label: 'Stage', icon: Play, color: 'bg-gradient-to-r from-amber-500 to-orange-500', gradient: 'gradient-warning' },
    { type: 'image', label: 'Send Image', icon: Image, color: 'bg-gradient-to-r from-blue-500 to-indigo-500', gradient: 'gradient-info' },
    { type: 'audio', label: 'Send Audio', icon: Mic, color: 'bg-gradient-to-r from-green-500 to-emerald-500', gradient: 'gradient-success' },
    { type: 'video', label: 'Send Video', icon: Video, color: 'bg-gradient-to-r from-purple-500 to-pink-500', gradient: 'gradient-purple' },
    { type: 'delay', label: 'Delay', icon: Clock, color: 'bg-gradient-to-r from-orange-500 to-red-500', gradient: 'gradient-warning' },
    { type: 'condition', label: 'Conditions', icon: GitBranch, color: 'bg-gradient-to-r from-violet-500 to-purple-500', gradient: 'gradient-purple' },
  ];

  return (
    <div className="min-h-screen bg-background">
      {/* Main Content */}
      <div className="flex h-screen bg-gradient-to-br from-slate-900 via-purple-900 to-slate-900">
        {/* Compact Tool Sidebar */}
        <div className="w-64 h-full bg-background/95 backdrop-blur-sm border-r border-primary/20 flex flex-col">
          <div className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-primary/20 scrollbar-track-transparent">
            <div className="p-3 pb-32 min-h-full">
              <h2 className="text-lg font-bold mb-3 text-foreground holographic-text">Flow Builder</h2>
              
              <div className="space-y-2 mb-4">
                 <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wide flex items-center gap-2">
                   <div className="w-1.5 h-1.5 bg-primary rounded-full pulse-glow"></div>
                   Add Nodes
                 </h3>
                 
                 <div className="grid gap-1.5">
                   {nodeTypeButtons.map((nodeType) => {
                     const IconComponent = nodeType.icon;
                     return (
                       <Button
                         key={nodeType.type}
                         onClick={() => addNode(nodeType.type)}
                         variant="ghost"
                         size="sm"
                         className={`w-full justify-start text-xs h-8 ${nodeType.color} text-white hover:opacity-90 transition-all duration-200 transform hover:scale-105 shadow-lg border-0`}
                       >
                         <IconComponent className="w-3.5 h-3.5 mr-2" />
                         {nodeType.label}
                       </Button>
                     );
                   })}
                 </div>
               </div>

              {/* Flow Controls */}
              <div className="space-y-2 mb-4">
                <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wide flex items-center gap-2">
                  <div className="w-1.5 h-1.5 bg-green-500 rounded-full pulse-glow"></div>
                  Flow Controls
                </h3>
                
                <div className="space-y-1.5">
                  <Input
                    placeholder="Flow name..."
                    value={flowName}
                    onChange={(e) => setFlowName(e.target.value)}
                    className="h-8 text-xs bg-background/50 border-border/50 focus:border-primary/50"
                  />
                </div>
              </div>

              {/* API Configuration */}
               <div className="space-y-2 mb-4">
                 <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wide flex items-center gap-2">
                   <div className="w-1.5 h-1.5 bg-orange-500 rounded-full pulse-glow"></div>
                   API Configuration
                 </h3>
                 
                 <div className="space-y-1.5">
                   <Input
                     placeholder="Field Instance..."
                     value={fieldInstance}
                     onChange={(e) => setFieldInstance(e.target.value)}
                     className="h-8 text-xs bg-background/50 border-border/50 focus:border-primary/50"
                   />
                   <Input
                     placeholder="OpenRouter API Key..."
                     type="password"
                     value={openRouterApiKey}
                     onChange={(e) => setOpenRouterApiKey(e.target.value)}
                     className="h-8 text-xs bg-background/50 border-border/50 focus:border-primary/50"
                   />
                 </div>
               </div>

               {/* Action Buttons */}
               <div className="space-y-2 mb-4">
                 <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wide flex items-center gap-2">
                   <div className="w-1.5 h-1.5 bg-purple-500 rounded-full pulse-glow"></div>
                   Actions
                 </h3>
                 
                 <div className="flex gap-1">
                   <Button
                     onClick={saveFlowToStorage}
                     size="sm"
                     className="flex-1 h-7 text-xs bg-gradient-to-r from-green-500 to-emerald-500 hover:from-green-600 hover:to-emerald-600 text-white border-0 shadow-lg"
                   >
                     <Save className="w-3 h-3 mr-1" />
                     Save
                   </Button>
                   
                   <Button
                     onClick={exportFlow}
                     size="sm"
                     variant="outline"
                     className="h-7 text-xs border-border/50 hover:bg-background/80"
                   >
                     <Download className="w-3 h-3" />
                   </Button>
                   
                   <Button
                     onClick={importFlow}
                     size="sm"
                     variant="outline"
                     className="h-7 text-xs border-border/50 hover:bg-background/80"
                   >
                     <Upload className="w-3 h-3" />
                   </Button>
                   
                   <Button
                     onClick={testFlow}
                     size="sm"
                     className="h-7 text-xs bg-gradient-to-r from-purple-500 to-violet-500 hover:from-purple-600 hover:to-violet-600 text-white border-0 shadow-lg"
                   >
                     <Play className="w-3 h-3" />
                   </Button>
                 </div>
               </div>

               {/* Flow Stats */}
               <div className="space-y-2 mb-20">
                 <h4 className="text-xs font-medium text-muted-foreground uppercase tracking-wide flex items-center gap-2">
                   <div className="w-1.5 h-1.5 bg-blue-500 rounded-full pulse-glow"></div>
                   Flow Stats
                 </h4>
                 <div className="text-xs text-muted-foreground space-y-1">
                   <div className="flex justify-between">
                     <span>Nodes:</span>
                     <span className="text-primary font-medium">{nodes.length}</span>
                   </div>
                   <div className="flex justify-between">
                     <span>Connections:</span>
                     <span className="text-primary font-medium">{edges.length}</span>
                   </div>
                 </div>
               </div>
            </div>
          </div>
        </div>

        {/* Flow Canvas */}
        <div className="flex-1 h-full relative bg-gradient-to-br from-background via-background/95 to-background/90">
          <div className="w-full h-full">
            <ReactFlow
              nodes={nodes}
              edges={edges}
              onNodesChange={onNodesChange}
              onEdgesChange={onEdgesChange}
              onConnect={onConnect}
              onEdgesDelete={onEdgesDelete}
              nodeTypes={nodeTypes}
              fitView
              fitViewOptions={{
                padding: 0.2,
                minZoom: 0.1,
                maxZoom: 2
              }}
              minZoom={0.05}
              maxZoom={4}
              defaultZoom={0.8}
              deleteKeyCode="Delete"
              className="bg-transparent"
              defaultEdgeOptions={{
                style: { stroke: 'hsl(var(--primary))', strokeWidth: 2 },
                markerEnd: { type: MarkerType.ArrowClosed, color: 'hsl(var(--primary))' },
              }}
              panOnScroll={true}
              selectionOnDrag={true}
              panOnDrag={[1, 2]}
              selectNodesOnDrag={false}
              nodesDraggable={true}
              nodesConnectable={true}
              elementsSelectable={true}
              snapToGrid={true}
              snapGrid={[15, 15]}
              connectionMode="loose"
              defaultViewport={{ x: 0, y: 0, zoom: 0.8 }}
              zoomOnScroll={true}
              zoomOnPinch={true}
              zoomOnDoubleClick={true}
              preventScrolling={false}
            >
              <Controls 
                className="!bg-background/95 !border-border/50 backdrop-blur rounded-lg shadow-xl"
                position="bottom-left"
                showZoom={true}
                showFitView={true}
                showInteractive={true}
                style={{ 
                  bottom: '20px', 
                  left: '20px', 
                  transform: 'scale(1.1)', 
                  transformOrigin: 'bottom left', 
                  zIndex: 1000,
                  visibility: 'visible',
                  opacity: 1,
                  backgroundColor: 'hsl(var(--background))',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '8px',
                  padding: '4px'
                }}
              />
              <MiniMap 
                className="!bg-background/95 !border-border/50 backdrop-blur rounded-lg shadow-xl"
                position="bottom-right"
                style={{ 
                  bottom: '20px', 
                  right: '20px', 
                  width: '150px', 
                  height: '100px', 
                  transform: 'scale(0.9)', 
                  transformOrigin: 'bottom right', 
                  zIndex: 1000,
                  visibility: 'visible',
                  opacity: 1
                }}
                nodeColor={(node) => {
                  switch (node.type) {
                    case 'start': return 'hsl(var(--primary))'
                    case 'message': return 'hsl(220 100% 60%)'
                    case 'condition': return 'hsl(45 100% 60%)'
                    case 'delay': return 'hsl(25 100% 60%)'
                    case 'image': return 'hsl(200 100% 60%)'
                    case 'audio': return 'hsl(120 100% 60%)'
                    case 'video': return 'hsl(280 100% 60%)'
                    case 'manual': return 'hsl(340 100% 60%)'
                    case 'prompt': return 'hsl(260 100% 60%)'
                    case 'stage': return 'hsl(35 100% 60%)'
                    default: return 'hsl(var(--muted-foreground))'
                  }
                }}
                maskColor="hsl(var(--muted) / 0.3)"
              />
              <Background 
                color="hsl(var(--primary) / 0.1)" 
                gap={20} 
                size={1}
                variant="dots"
              />
            </ReactFlow>
          </div>
          
          {/* Test Modal */}
          {showTestModal && (
            <div className="absolute inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50">
              <Card className="w-96 max-h-96 overflow-y-auto bg-card/95 backdrop-blur-xl border-border/50 futuristic-border">
                <div className="p-6">
                  <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-semibold holographic-text">Test Flow</h3>
                    <Button
                      onClick={() => setShowTestModal(false)}
                      variant="ghost"
                      size="sm"
                      className="hover:bg-destructive/20 hover:text-destructive transition-colors"
                    >
                      <X className="w-4 h-4" />
                    </Button>
                  </div>
                  
                  <div className="space-y-3 mb-4 max-h-48 overflow-y-auto">
                    {testMessages.map((msg, index) => (
                      <div key={index} className={`flex ${msg.type === 'user' ? 'justify-end' : 'justify-start'}`}>
                        <div className={`max-w-xs p-2 rounded-lg text-sm ${
                          msg.type === 'user' 
                            ? 'bg-primary text-primary-foreground' 
                            : 'bg-muted text-muted-foreground'
                        }`}>
                          {msg.type === 'bot' && <Bot className="w-3 h-3 inline mr-1" />}
                          {msg.content}
                        </div>
                      </div>
                    ))}
                  </div>
                  
                  <div className="flex gap-2">
                    <Input
                      placeholder="Type a message..."
                      value={testInput}
                      onChange={(e) => setTestInput(e.target.value)}
                      onKeyPress={(e) => e.key === 'Enter' && sendTestMessage()}
                      className="flex-1"
                    />
                    <Button onClick={sendTestMessage} size="sm">
                      <Send className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
              </Card>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}