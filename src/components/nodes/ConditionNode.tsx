import React, { useState } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { GitBranch, Edit3 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function ConditionNode({ data, id }: NodeProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [condition, setCondition] = useState((data?.condition as string) || 'user_input contains "yes"');

  const handleSave = () => {
    setIsEditing(false);
    // In a real app, you'd update the node data here
  };

  return (
    <div className="bg-card rounded-lg shadow-node border border-border min-w-[250px] max-w-[350px]">
      <Handle 
        type="target" 
        position={Position.Top} 
        className="w-3 h-3 bg-primary border-2 border-white"
      />
      
      <div className="p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center">
            <div className="w-3 h-3 rounded-full bg-node-condition mr-2" />
            <GitBranch className="w-4 h-4 text-node-condition mr-2" />
            <span className="text-sm font-medium text-foreground">Condition</span>
          </div>
          <Button
            size="sm"
            variant="ghost"
            onClick={() => setIsEditing(!isEditing)}
            className="h-6 w-6 p-0"
          >
            <Edit3 className="w-3 h-3" />
          </Button>
        </div>
        
        {isEditing ? (
          <div className="space-y-2">
            <Input
              value={condition}
              onChange={(e) => setCondition(e.target.value)}
              className="text-sm"
              placeholder="user_input contains..."
            />
            <Button size="sm" onClick={handleSave} className="w-full">
              Save
            </Button>
          </div>
        ) : (
          <div className="bg-muted/50 rounded p-3 text-sm text-foreground font-mono">
            {condition}
          </div>
        )}
      </div>
      
      {/* True/False outputs */}
      <div className="flex justify-between px-4 pb-2">
        <div className="text-xs text-green-500 font-medium">TRUE</div>
        <div className="text-xs text-red-500 font-medium">FALSE</div>
      </div>
      
      <Handle 
        id="true"
        type="source" 
        position={Position.Bottom} 
        style={{ left: '25%' }}
        className="w-3 h-3 bg-green-500 border-2 border-white"
      />
      <Handle 
        id="false"
        type="source" 
        position={Position.Bottom} 
        style={{ left: '75%' }}
        className="w-3 h-3 bg-red-500 border-2 border-white"
      />
    </div>
  );
}