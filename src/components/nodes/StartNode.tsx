import React from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Play } from 'lucide-react';

export default function StartNode({ data, id }: NodeProps) {
  return (
    <div className="bg-gradient-to-br from-green-500 to-emerald-600 rounded-lg shadow-node border border-border/50 min-w-[150px] text-center futuristic-border glow-on-hover transition-all duration-300 relative overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/10 to-transparent opacity-0 hover:opacity-100 transition-opacity duration-500"></div>
      <div className="p-4 relative z-10">
        <div className="flex items-center justify-center mb-2">
          <Play className="w-5 h-5 text-white mr-2 drop-shadow-lg" />
          <span className="text-sm font-semibold text-white drop-shadow-lg holographic-text">Start</span>
        </div>
        <div className="text-xs text-white/90 drop-shadow">
          Flow begins here
        </div>
      </div>
      
      <Handle 
        type="source" 
        position={Position.Bottom} 
        id="start-output"
        className="w-3 h-3 bg-white border-2 border-green-500 shadow-lg"
      />
    </div>
  );
}