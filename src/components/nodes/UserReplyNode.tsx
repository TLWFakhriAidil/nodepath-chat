import React from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { MessageCircle, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';

/**
 * UserReplyNode - A flow node that waits indefinitely for user input before proceeding
 * This node will pause the conversation flow until the user provides a response
 */
export default function UserReplyNode({ data, id }: NodeProps) {

  return (
    <div className="bg-card rounded-lg shadow-node border border-border min-w-[250px] max-w-[350px]">
      <Handle 
        type="target" 
        position={Position.Top} 
        id="user-reply-input"
        className="w-3 h-3 bg-primary border-2 border-white"
      />
      
      <div className="p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center">
            <div className="w-3 h-3 rounded-full bg-green-500 mr-2" />
            <MessageCircle className="w-4 h-4 text-green-600 mr-2" />
            <span className="text-sm font-medium text-black">User Reply</span>
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
          <div className="bg-green-50 rounded p-3 text-sm">
            <div className="text-xs text-muted-foreground mb-1">User Reply Node:</div>
            <div className="text-black">
              This node will wait indefinitely for the user to reply before proceeding to the next step in the flow.
            </div>
          </div>
          <div className="text-xs text-muted-foreground italic">
            ⏳ Waits until user provides input
          </div>
        </div>
      </div>
      
      <Handle 
        type="source" 
        position={Position.Bottom} 
        id="user-reply-output"
        className="w-3 h-3 bg-primary border-2 border-white"
      />
    </div>
  );
}