import React from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  Node,
  Edge,
  NodeTypes,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { MessageSquare, Image, Mic, Video, Clock, GitBranch, Play } from 'lucide-react';
import { ChatbotFlow } from '@/types/chatbot';

// Read-only preview node component
const PreviewNode = ({ data, type }: { data: any; type: string }) => {
  const getIcon = () => {
    switch (type) {
      case 'start': return <Play className="w-4 h-4" />;
      case 'message': return <MessageSquare className="w-4 h-4" />;
      case 'image': return <Image className="w-4 h-4" />;
      case 'audio': return <Mic className="w-4 h-4" />;
      case 'video': return <Video className="w-4 h-4" />;
      case 'delay': return <Clock className="w-4 h-4" />;
      case 'condition': return <GitBranch className="w-4 h-4" />;
      default: return <MessageSquare className="w-4 h-4" />;
    }
  };

  const getTypeColor = () => {
    switch (type) {
      case 'start': return 'bg-green-100 border-green-300 text-green-800';
      case 'message': return 'bg-blue-100 border-blue-300 text-blue-800';
      case 'image': return 'bg-purple-100 border-purple-300 text-purple-800';
      case 'audio': return 'bg-orange-100 border-orange-300 text-orange-800';
      case 'video': return 'bg-pink-100 border-pink-300 text-pink-800';
      case 'delay': return 'bg-yellow-100 border-yellow-300 text-yellow-800';
      case 'condition': return 'bg-red-100 border-red-300 text-red-800';
      default: return 'bg-gray-100 border-gray-300 text-gray-800';
    }
  };

  const getContent = () => {
    switch (type) {
      case 'message':
        return data.message || 'No message set';
      case 'image':
        return data.imageUrl || 'No image set';
      case 'audio':
        return `Audio (${data.duration || 30}s)`;
      case 'video':
        return `Video (${data.duration || 60}s)`;
      case 'delay':
        return `Wait ${data.delay || data.delaySeconds || 5}s`;
      case 'condition':
        return data.condition || 'No condition set';
      default:
        return data.label || 'Start';
    }
  };

  return (
    <div className={`p-3 rounded-lg border-2 min-w-[200px] max-w-[250px] ${getTypeColor()}`}>
      <div className="flex items-center gap-2 mb-2">
        {getIcon()}
        <Badge variant="secondary" className="text-xs">
          {type.charAt(0).toUpperCase() + type.slice(1)}
        </Badge>
      </div>
      <div className="text-sm font-medium mb-1">{data.label}</div>
      <div className="text-xs opacity-75 break-words">
        {getContent()}
      </div>
    </div>
  );
};

const previewNodeTypes: NodeTypes = {
  start: (props) => <PreviewNode {...props} type="start" />,
  message: (props) => <PreviewNode {...props} type="message" />,
  image: (props) => <PreviewNode {...props} type="image" />,
  audio: (props) => <PreviewNode {...props} type="audio" />,
  video: (props) => <PreviewNode {...props} type="video" />,
  delay: (props) => <PreviewNode {...props} type="delay" />,
  condition: (props) => <PreviewNode {...props} type="condition" />,
};

interface FlowPreviewProps {
  flow: ChatbotFlow | null;
}

export default function FlowPreview({ flow }: FlowPreviewProps) {
  if (!flow) {
    return (
      <Card className="flex-1 p-8 bg-muted/20">
        <div className="text-center text-muted-foreground">
          <MessageSquare className="w-12 h-12 mx-auto mb-4 opacity-50" />
          <h3 className="text-lg font-medium mb-2">No Flow Selected</h3>
          <p>Select a flow from the sidebar to preview its structure</p>
        </div>
      </Card>
    );
  }

  const nodes: Node[] = flow.nodes.map(node => ({
    id: node.id,
    type: node.type,
    position: node.position,
    data: node.data,
    draggable: false,
    selectable: false,
  }));

  const edges: Edge[] = flow.edges.map(edge => ({
    id: edge.id,
    source: edge.source,
    target: edge.target,
    style: { stroke: 'hsl(var(--primary))', strokeWidth: 2 },
    animated: true,
  }));

  return (
    <div className="flex-1 flex flex-col">
      {/* Flow Metadata */}
      <Card className="p-4 mb-4">
        <div className="flex justify-between items-start">
          <div>
            <h2 className="text-xl font-semibold mb-1">{flow.name}</h2>
            <p className="text-sm text-muted-foreground mb-2">{flow.description}</p>
            <div className="flex gap-4 text-xs text-muted-foreground">
              <span>Nodes: {flow.nodes.length}</span>
              <span>Connections: {flow.edges.length}</span>
              <span>Created: {new Date(flow.createdAt).toLocaleDateString()}</span>
              <span>Updated: {new Date(flow.updatedAt).toLocaleDateString()}</span>
            </div>
          </div>
          <Badge variant="outline">Preview Mode</Badge>
        </div>
      </Card>

      {/* Flow Visualization */}
      <Card className="flex-1 relative overflow-hidden">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          nodeTypes={previewNodeTypes}
          fitView
          style={{ backgroundColor: 'hsl(var(--background))' }}
          nodesDraggable={false}
          nodesConnectable={false}
          elementsSelectable={false}
          panOnDrag={true}
          zoomOnScroll={true}
          zoomOnPinch={true}
        >
          <Controls 
            className="bg-card border-border"
            showInteractive={false}
          />
          <MiniMap 
            className="bg-card border-border"
            style={{ backgroundColor: 'hsl(var(--card))' }}
            maskColor="hsl(var(--muted) / 0.3)"
            nodeStrokeWidth={2}
          />
          <Background 
            color="hsl(var(--border))" 
            gap={20} 
            style={{ backgroundColor: 'hsl(var(--background))' }}
          />
        </ReactFlow>
      </Card>
    </div>
  );
}