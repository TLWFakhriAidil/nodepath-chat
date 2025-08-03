import React, { useState } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Settings, Edit3, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';

export default function ManualNode({ data, id }: NodeProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [expectedInput, setExpectedInput] = useState((data?.expectedInput as string) || '');
  const [responseOutput, setResponseOutput] = useState((data?.responseOutput as string) || '');

  const handleSave = () => {
    setIsEditing(false);
    // Update the node data through the parent
    if (data?.onUpdate) {
      (data.onUpdate as Function)(id, { 
        expectedInput,
        responseOutput
      });
    }
  };

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
          <div className="space-y-3">
            <div className="space-y-2">
              <label className="text-xs font-medium text-muted-foreground">Expected User Input:</label>
              <Input
                value={expectedInput}
                onChange={(e) => setExpectedInput(e.target.value)}
                placeholder="e.g., 'yes', 'hello', 'I want info'..."
                className="text-sm"
              />
            </div>
            
            <div className="space-y-2">
              <label className="text-xs font-medium text-muted-foreground">Bot Response:</label>
              <Textarea
                value={responseOutput}
                onChange={(e) => setResponseOutput(e.target.value)}
                placeholder="Response to send when input matches..."
                className="text-sm min-h-[60px]"
              />
            </div>
            
            <Button size="sm" onClick={handleSave} className="w-full">
              Save
            </Button>
          </div>
        ) : (
          <div className="space-y-2">
            <div className="bg-blue-50 rounded p-3 text-sm">
              <div className="text-xs text-muted-foreground mb-1">Expected:</div>
              <div className="text-black font-medium">
                {expectedInput || 'No input pattern set'}
              </div>
            </div>
            <div className="bg-muted/50 rounded p-3 text-sm">
              <div className="text-xs text-muted-foreground mb-1">Response:</div>
              <div className="text-black">
                {responseOutput || 'No response set'}
              </div>
            </div>
          </div>
        )}
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