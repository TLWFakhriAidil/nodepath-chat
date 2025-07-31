import React, { useState } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Clock, Edit3, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function DelayNode({ data, id }: NodeProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [delay, setDelay] = useState((data?.delay as number) || 5);

  const handleSave = () => {
    setIsEditing(false);
    if (data?.onUpdate) {
      (data.onUpdate as Function)(id, { delay, delaySeconds: delay });
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
            <div className="w-3 h-3 rounded-full bg-node-delay mr-2" />
            <Clock className="w-4 h-4 text-node-delay mr-2" />
            <span className="text-sm font-medium text-black">Delay</span>
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
                value={delay}
                onChange={(e) => setDelay(parseInt(e.target.value) || 0)}
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
          <div className="bg-muted/50 rounded p-3 text-center">
            <div className="text-2xl font-bold text-node-delay">{delay}</div>
            <div className="text-xs text-muted-foreground">seconds</div>
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