import React, { useCallback, useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
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
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { MessageSquare, GitBranch, Clock, Play, Download, Image, Mic, Video, Save } from 'lucide-react';
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

const nodeTypes: NodeTypes = {
  message: MessageNode,
  condition: ConditionNode,
  delay: DelayNode,
  start: StartNode,
  image: ImageNode,
  audio: AudioNode,
  video: VideoNode,
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

export default function ChatbotBuilder({ onTestFlow }: { onTestFlow?: (flowId: string) => void }) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [selectedNodeType, setSelectedNodeType] = useState<string | null>(null);
  const [flowName, setFlowName] = useState('');
  const [currentFlowId, setCurrentFlowId] = useState<string | null>(null);
  const [toolkitMode, setToolkitMode] = useState<'manual' | 'prompt'>('manual');
  const [flowPrompt, setFlowPrompt] = useState('');
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

  // Load flow for editing if edit parameter is present
  useEffect(() => {
    const loadFlowForEdit = async () => {
      const editFlowId = searchParams.get('edit');
      if (editFlowId) {
        const flowToEdit = await getFlow(editFlowId);
        if (flowToEdit) {
          setFlowName(flowToEdit.name);
          setCurrentFlowId(flowToEdit.id);
          setToolkitMode(flowToEdit.toolkitMode || 'manual');
          setFlowPrompt(flowToEdit.flowPrompt || '');
          setNodes(flowToEdit.nodes.map(node => ({
            ...node,
            data: { ...node.data, onDelete: deleteNode, onUpdate: updateNodeData }
          })));
          setEdges(flowToEdit.edges);
          toast({
            title: "Flow loaded for editing",
            description: `"${flowToEdit.name}" is now loaded in the editor`
          });
        }
      }
    };
    loadFlowForEdit();
  }, [searchParams, setNodes, setEdges, toast, deleteNode]);

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
                 'New Video',
          message: type === 'message' ? 'Enter your message here...' : undefined,
          condition: type === 'condition' ? 'user_input contains "yes"' : undefined,
          delay: type === 'delay' ? 5 : undefined,
          imageUrl: type === 'image' ? '' : undefined,
          caption: type === 'image' || type === 'video' ? 'Caption...' : undefined,
          audioUrl: type === 'audio' ? '' : undefined,
          videoUrl: type === 'video' ? '' : undefined,
          duration: type === 'audio' || type === 'video' ? (type === 'audio' ? 30 : 60) : undefined,
          onDelete: deleteNode,
          onUpdate: updateNodeData,
        },
      };
      setNodes((nds) => nds.concat(newNode));
    },
    [setNodes, deleteNode, updateNodeData]
  );

  const saveFlowToStorage = useCallback(async () => {
    console.log('=== SAVE FLOW DEBUG ===');
    console.log('Flow name:', flowName);
    console.log('Current nodes state:', nodes);
    console.log('Current edges state:', edges);
    console.log('Nodes length:', nodes.length);
    console.log('Edges length:', edges.length);
    
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

    console.log('Saving flow with nodes:', nodes);
    console.log('Saving flow with edges:', edges);

    const flowData: ChatbotFlow = {
      id: currentFlowId || `flow_${Date.now()}_${Math.random().toString(36).substring(2)}`,
      name: flowName,
      description: `Chatbot flow: ${flowName}`,
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
      updatedAt: new Date().toISOString(),
      toolkitMode: toolkitMode,
      flowPrompt: toolkitMode === 'prompt' ? flowPrompt : undefined
    };

    console.log('Flow data to save:', flowData);
    await saveFlow(flowData);
    setCurrentFlowId(flowData.id);
    
    console.log('Flow saved.');
    
    toast({
      title: "Flow saved",
      description: `"${flowName}" has been saved successfully`
    });
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

  const testFlow = useCallback(() => {
    if (!currentFlowId) {
      saveFlowToStorage();
      return;
    }
    onTestFlow?.(currentFlowId);
    navigate('/test');
  }, [currentFlowId, saveFlowToStorage, onTestFlow, navigate]);

  const nodeTypeButtons = [
    { type: 'message', label: 'Send Message', icon: MessageSquare, color: 'bg-node-message' },
    { type: 'image', label: 'Send Image', icon: Image, color: 'bg-blue-500' },
    { type: 'audio', label: 'Send Audio', icon: Mic, color: 'bg-green-500' },
    { type: 'video', label: 'Send Video', icon: Video, color: 'bg-purple-500' },
    { type: 'delay', label: 'Delay', icon: Clock, color: 'bg-node-delay' },
    { type: 'condition', label: 'Conditions', icon: GitBranch, color: 'bg-node-condition' },
  ];

  return (
    <div className="h-screen bg-background flex">
      {/* Sidebar */}
      <Card className="w-80 bg-card border-border rounded-none border-r">
        <div className="p-6">
          <h2 className="text-xl font-semibold mb-6 text-foreground">Chatbot Builder</h2>
          
          <div className="space-y-4 mb-8">
            <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
              Add Nodes
            </h3>
            {nodeTypeButtons.map(({ type, label, icon: Icon, color }) => (
              <Button
                key={type}
                onClick={() => addNode(type)}
                variant="outline"
                className="w-full justify-start h-12 border-border"
              >
                <div className={`w-3 h-3 rounded-full ${color} mr-3`} />
                <Icon className="w-4 h-4 mr-2" />
                {label}
              </Button>
            ))}
          </div>

          <div className="space-y-3">
            <Input
              placeholder="Enter flow name..."
              value={flowName}
              onChange={(e) => setFlowName(e.target.value)}
              className="w-full"
            />
            
            {/* Toolkit Mode Section */}
            <div className="space-y-3 p-4 bg-muted/30 rounded-lg border">
              <h3 className="text-sm font-medium text-foreground">Toolkit Mode</h3>
              <div className="space-y-2">
                <label className="flex items-center space-x-2">
                  <input
                    type="radio"
                    name="toolkitMode"
                    value="manual"
                    checked={toolkitMode === 'manual'}
                    onChange={(e) => setToolkitMode(e.target.value as 'manual' | 'prompt')}
                    className="form-radio"
                  />
                  <span className="text-sm">Manual Mode</span>
                </label>
                <label className="flex items-center space-x-2">
                  <input
                    type="radio"
                    name="toolkitMode"
                    value="prompt"
                    checked={toolkitMode === 'prompt'}
                    onChange={(e) => setToolkitMode(e.target.value as 'manual' | 'prompt')}
                    className="form-radio"
                  />
                  <span className="text-sm">Prompt Mode</span>
                </label>
              </div>
              
              {toolkitMode === 'prompt' && (
                <textarea
                  placeholder="Enter AI prompt for this flow..."
                  value={flowPrompt}
                  onChange={(e) => setFlowPrompt(e.target.value)}
                  className="w-full p-2 text-sm border rounded-md bg-background"
                  rows={3}
                />
              )}
              
              <div className="text-xs text-muted-foreground">
                {toolkitMode === 'manual' 
                  ? 'Use predefined responses step-by-step'
                  : 'AI generates dynamic responses using your prompt'
                }
              </div>
            </div>
            
            <Button 
              onClick={saveFlowToStorage}
              variant="default"
              className="w-full"
            >
              <Save className="w-4 h-4 mr-2" />
              Save Flow
            </Button>
            
            <Button 
              onClick={exportFlow}
              variant="secondary" 
              className="w-full"
            >
              <Download className="w-4 h-4 mr-2" />
              Export JSON
            </Button>
            
            <Button 
              onClick={testFlow}
              variant="default" 
              className="w-full bg-gradient-primary"
            >
              <Play className="w-4 h-4 mr-2" />
              Test Flow
            </Button>
          </div>

          <div className="mt-8 p-4 bg-muted/50 rounded-lg">
            <h4 className="text-sm font-medium mb-2">Flow Stats</h4>
            <div className="text-sm text-muted-foreground space-y-1">
              <div>Nodes: {nodes.length}</div>
              <div>Connections: {edges.length}</div>
            </div>
          </div>
        </div>
      </Card>

      {/* Flow Canvas */}
      <div className="flex-1 relative">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onEdgesDelete={onEdgesDelete}
          nodeTypes={nodeTypes}
          fitView
          deleteKeyCode="Delete"
          style={{ backgroundColor: 'hsl(var(--flow-background))' }}
          defaultEdgeOptions={{
            style: { stroke: 'hsl(var(--primary))', strokeWidth: 2 },
            markerEnd: { type: MarkerType.ArrowClosed, color: 'hsl(var(--primary))' },
          }}
        >
          <Controls 
            className="bg-card border-border"
          />
          <MiniMap 
            className="bg-card border-border"
            style={{ backgroundColor: 'hsl(var(--card))' }}
            maskColor="hsl(var(--muted) / 0.3)"
          />
          <Background 
            color="hsl(var(--border))" 
            gap={20} 
            style={{ backgroundColor: 'hsl(var(--flow-background))' }}
          />
        </ReactFlow>
      </div>
    </div>
  );
}