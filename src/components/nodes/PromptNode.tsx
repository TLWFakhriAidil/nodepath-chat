import React, { useState, useEffect } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Sparkles, Trash2, Edit3, Check, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';

export default function PromptNode({ data, id }: NodeProps) {
  
  const [isEditing, setIsEditing] = useState(false);
  const [systemPrompt, setSystemPrompt] = useState(String(data.systemPrompt || ''));

  useEffect(() => {
    // Initialize local state with node data
    setSystemPrompt(String(data.systemPrompt || ''));
  }, [data]);


  const handleSave = () => {
    const updatedData = {
      systemPrompt,
      node_type: 'ai_prompt'
    };
    
    if (data?.onUpdate && typeof data.onUpdate === 'function') {
      data.onUpdate(id, updatedData);
    }
    
    setIsEditing(false);
  };

  const handleCancel = () => {
    // Reset to original values
    setSystemPrompt(String(data.systemPrompt || ''));
    setIsEditing(false);
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
            {isEditing ? (
              <>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={handleSave}
                  className="h-6 w-6 p-0 text-green-600 hover:text-green-700"
                >
                  <Check className="w-3 h-3" />
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={handleCancel}
                  className="h-6 w-6 p-0 text-muted-foreground hover:text-foreground"
                >
                  <X className="w-3 h-3" />
                </Button>
              </>
            ) : (
              <Button
                size="sm"
                variant="ghost"
                onClick={() => setIsEditing(true)}
                className="h-6 w-6 p-0 text-muted-foreground hover:text-foreground"
              >
                <Edit3 className="w-3 h-3" />
              </Button>
            )}
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
        
        <div className="space-y-3">
          {isEditing ? (
            <>
              <div>
                <label className="text-xs text-muted-foreground mb-1 block">System Prompt:</label>
                <Textarea
                  value={String(systemPrompt || '')}
                  onChange={(e) => setSystemPrompt(e.target.value)}
                  placeholder="You are a helpful assistant that responds clearly and concisely."
                  className="min-h-[80px] text-xs resize-none text-black bg-white border border-gray-300"
                />
              </div>
            </>
          ) : (
            <>
              <div className="bg-purple-50 rounded p-3 text-sm">
                <div className="text-xs text-muted-foreground mb-1">System Prompt:</div>
                <div className="text-black">
                  {(() => {
                    const prompt = (systemPrompt || data.systemPrompt) as string;
                    if (prompt) {
                      return prompt.length > 100 ? `${prompt.substring(0, 100)}...` : prompt;
                    }
                    return 'No system prompt configured';
                  })()}
                </div>
              </div>
              <div className="text-xs text-muted-foreground italic">
                🤖 Click edit to configure this AI prompt node
              </div>
            </>
          )}
        </div>
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