import React from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Settings, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';

export default function ManualNode({ data, id }: NodeProps) {

  return (
    <div className="bg-card rounded-lg shadow-node border border-border min-w-[250px] max-w-[350px]">
      <Handle 
        type="target" 
        position={Position.Top} 
        id="manual-input"
        className="w-3 h-3 bg-primary border-2 border-white"
      />
      
      <div className="p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center">
            <div className="w-3 h-3 rounded-full bg-blue-500 mr-2" />
            <Settings className="w-4 h-4 text-blue-600 mr-2" />
            <span className="text-sm font-medium text-black">Manual</span>
          </div>
          <div className="flex items-center gap-1">
            <Button
              size="sm"
              variant="ghost"
              onClick={() => (data?.onDelete as Function)?.(id)}
              className="h-6 w-6 p-0 text-destructive hover:text-destructive"
            >
              <Trash2 className="w-3 h-3" />
            </Button>
          </div>
        </div>
        
        <div className="space-y-2">
          <div className="bg-blue-50 rounded p-3 text-sm">
            <div className="text-xs text-muted-foreground mb-1">Manual Response Node:</div>
            <div className="text-black">
              This node will pause the conversation and wait for manual intervention from a staff member.
            </div>
          </div>
          <div className="text-xs text-muted-foreground italic">
            👤 Staff will handle responses manually
          </div>
        </div>
      </div>
      
      <Handle 
        type="source" 
        position={Position.Bottom} 
        id="manual-output"
        className="w-3 h-3 bg-primary border-2 border-white"
      />
    </div>
  );
}