import React, { useState } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Timer, Edit3, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

/**
 * WaitingReplyTimesNode - A flow node that waits for a specified number of seconds for user input
 * If no input is received within the timeout, the flow continues to the next node
 */
export default function WaitingReplyTimesNode({ data, id }: NodeProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [waitTime, setWaitTime] = useState((data?.waitTime as number) || 5);

  const handleSave = () => {
    setIsEditing(false);
    if (data?.onUpdate) {
      (data.onUpdate as Function)(id, { waitTime, waitTimeSeconds: waitTime });
    }
  };

  return (
    <div className="bg-card rounded-lg shadow-node border border-border min-w-[200px]">
      <Handle 
        type="target" 
        position={Position.Top} 
        className="w-3 h-3 bg-primary border-2 border-white"
      />
      
      <div className="p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center">
            <div className="w-3 h-3 rounded-full bg-orange-500 mr-2" />
            <Timer className="w-4 h-4 text-orange-600 mr-2" />
            <span className="text-sm font-medium text-black">Wait Reply</span>
          </div>
          <div className="flex items-center gap-1">
            <Button
              size="sm"
              variant="ghost"
              onClick={() => setIsEditing(!isEditing)}
              className="h-6 w-6 p-0"
            >
              <Edit3 className="w-3 h-3" />
            </Button>
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
        
        {isEditing ? (
          <div className="space-y-2">
            <div className="flex items-center space-x-2">
              <Input
                type="number"
                value={waitTime}
                onChange={(e) => setWaitTime(parseInt(e.target.value) || 0)}
                className="text-sm text-black bg-white border-gray-300"
                min="1"
                max="3600"
              />
              <span className="text-sm text-black">seconds</span>
            </div>
            <Button size="sm" onClick={handleSave} className="w-full">
              Save
            </Button>
          </div>
        ) : (
          <div className="space-y-2">
            <div className="bg-muted/50 rounded p-3 text-center">
              <div className="text-2xl font-bold text-orange-600">{waitTime}</div>
              <div className="text-xs text-muted-foreground">seconds</div>
            </div>
            <div className="bg-orange-50 rounded p-2 text-xs">
              <div className="text-muted-foreground mb-1">Wait for user reply:</div>
              <div className="text-black">
                Waits {waitTime}s for user input, then continues if no reply
              </div>
            </div>
          </div>
        )}
      </div>
      
      <Handle 
        type="source" 
        position={Position.Bottom} 
        className="w-3 h-3 bg-primary border-2 border-white"
      />
    </div>
  );
}