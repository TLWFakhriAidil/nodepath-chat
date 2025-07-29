import React from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Play } from 'lucide-react';

export default function StartNode({ data }: NodeProps) {
  return (
    <div className="bg-gradient-success rounded-lg shadow-node border-2 border-white/20 min-w-[120px]">
      <div className="p-4 text-center">
        <div className="flex items-center justify-center mb-2">
          <Play className="w-5 h-5 text-white" />
        </div>
        <div className="text-white font-medium text-sm">Start</div>
      </div>
      <Handle 
        type="source" 
        position={Position.Bottom} 
        id="start-output"
        className="w-3 h-3 bg-white border-2 border-node-start"
      />
    </div>
  );
}