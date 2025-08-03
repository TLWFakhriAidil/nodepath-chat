import React, { useState } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Sparkles, Edit3, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';

export default function PromptNode({ data, id }: NodeProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [nodePrompt, setNodePrompt] = useState((data?.nodePrompt as string) || '');

  const handleSave = () => {
    setIsEditing(false);
    // Update the node data through the parent
    if (data?.onUpdate) {
      (data.onUpdate as Function)(id, { 
        nodePrompt
      });
    }
  };

  return (
    <div className="bg-card rounded-lg shadow-node border border-border min-w-[250px] max-w-[350px]">
      <Handle 
        type="target" 
        position={Position.Top} 
        id="prompt-input"
        className="w-3 h-3 bg-primary border-2 border-white"
      />
      
      <div className="p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center">
            <div className="w-3 h-3 rounded-full bg-purple-500 mr-2" />
            <Sparkles className="w-4 h-4 text-purple-600 mr-2" />
            <span className="text-sm font-medium text-black">AI Prompt</span>
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
              <label className="text-xs font-medium text-muted-foreground">AI Prompt:</label>
              <Textarea
                value={nodePrompt}
                onChange={(e) => setNodePrompt(e.target.value)}
                placeholder="Enter the AI prompt for generating responses at this step..."
                className="text-sm min-h-[100px]"
              />
              <div className="text-xs text-muted-foreground">
                The AI will use this prompt along with the user's input to generate a dynamic response.
              </div>
            </div>
            
            <Button size="sm" onClick={handleSave} className="w-full">
              Save
            </Button>
          </div>
        ) : (
          <div className="space-y-2">
            <div className="bg-purple-50 rounded p-3 text-sm">
              <div className="text-xs text-muted-foreground mb-1">AI Prompt:</div>
              <div className="text-black">
                {nodePrompt ? (
                  nodePrompt.length > 100 ? 
                    `${nodePrompt.substring(0, 100)}...` : 
                    nodePrompt
                ) : (
                  'No prompt set'
                )}
              </div>
            </div>
            <div className="text-xs text-muted-foreground italic">
              🤖 AI will generate responses dynamically
            </div>
          </div>
        )}
      </div>
      
      <Handle 
        type="source" 
        position={Position.Bottom} 
        id="prompt-output"
        className="w-3 h-3 bg-primary border-2 border-white"
      />
    </div>
  );
}