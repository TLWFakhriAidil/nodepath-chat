import React, { useCallback, useState } from 'react';
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
import { MessageSquare, GitBranch, Clock, Play, Download, Image, Mic, Video } from 'lucide-react';

import MessageNode from './nodes/MessageNode';
import ConditionNode from './nodes/ConditionNode';
import DelayNode from './nodes/DelayNode';
import StartNode from './nodes/StartNode';
import ImageNode from './nodes/ImageNode';
import AudioNode from './nodes/AudioNode';
import VideoNode from './nodes/VideoNode';

const nodeTypes: NodeTypes = {
  messageNode: MessageNode,
  conditionNode: ConditionNode,
  delayNode: DelayNode,
  startNode: StartNode,
  imageNode: ImageNode,
  audioNode: AudioNode,
  videoNode: VideoNode,
};

const initialNodes: Node[] = [
  {
    id: 'start-1',
    type: 'startNode',
    position: { x: 250, y: 100 },
    data: { label: 'Start' },
  },
];

const initialEdges: Edge[] = [];

export default function ChatbotBuilder() {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [selectedNodeType, setSelectedNodeType] = useState<string | null>(null);

  const onConnect = useCallback(
    (params: Edge | Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  );

  const deleteNode = useCallback(
    (nodeId: string) => {
      setNodes((nds) => nds.filter((node) => node.id !== nodeId));
      setEdges((eds) => eds.filter((edge) => edge.source !== nodeId && edge.target !== nodeId));
    },
    [setNodes, setEdges]
  );

  const addNode = useCallback(
    (type: string) => {
      const newNode: Node = {
        id: `${type}-${Date.now()}`,
        type,
        position: { x: Math.random() * 400 + 100, y: Math.random() * 400 + 200 },
        data: {
          label: type === 'messageNode' ? 'New Message' : 
                 type === 'conditionNode' ? 'New Condition' : 
                 type === 'delayNode' ? 'New Delay' :
                 type === 'imageNode' ? 'New Image' :
                 type === 'audioNode' ? 'New Audio' :
                 'New Video',
          message: type === 'messageNode' ? 'Enter your message here...' : undefined,
          condition: type === 'conditionNode' ? 'user_input contains "yes"' : undefined,
          delay: type === 'delayNode' ? 5 : undefined,
          imageUrl: type === 'imageNode' ? '' : undefined,
          caption: type === 'imageNode' || type === 'videoNode' ? 'Caption...' : undefined,
          audioUrl: type === 'audioNode' ? '' : undefined,
          videoUrl: type === 'videoNode' ? '' : undefined,
          duration: type === 'audioNode' || type === 'videoNode' ? (type === 'audioNode' ? 30 : 60) : undefined,
          onDelete: deleteNode,
        },
      };
      setNodes((nds) => nds.concat(newNode));
    },
    [setNodes, deleteNode]
  );

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

  const nodeTypeButtons = [
    { type: 'messageNode', label: 'Send Message', icon: MessageSquare, color: 'bg-node-message' },
    { type: 'imageNode', label: 'Send Image', icon: Image, color: 'bg-blue-500' },
    { type: 'audioNode', label: 'Send Audio', icon: Mic, color: 'bg-green-500' },
    { type: 'videoNode', label: 'Send Video', icon: Video, color: 'bg-purple-500' },
    { type: 'delayNode', label: 'Delay', icon: Clock, color: 'bg-node-delay' },
    { type: 'conditionNode', label: 'Conditions', icon: GitBranch, color: 'bg-node-condition' },
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
            <Button 
              onClick={exportFlow}
              variant="secondary" 
              className="w-full"
            >
              <Download className="w-4 h-4 mr-2" />
              Export Flow
            </Button>
            
            <Button 
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
          nodeTypes={nodeTypes}
          fitView
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