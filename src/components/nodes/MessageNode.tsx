import React, { useState } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { MessageSquare, Edit3, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Input } from '@/components/ui/input';

export default function MessageNode({ data, id }: NodeProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [message, setMessage] = useState((data?.message as string) || 'Enter your message here...');
  const [expectedInput, setExpectedInput] = useState((data?.expectedInput as string) || '');
  const [responseOutput, setResponseOutput] = useState((data?.responseOutput as string) || '');

  const handleSave = () => {
    setIsEditing(false);
    // Update the node data through the parent
    if (data?.onUpdate) {
      (data.onUpdate as Function)(id, { 
        message,
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
        id="message-input"
        className="w-3 h-3 bg-primary border-2 border-white"
      />
      
      <div className="p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center">
            <div className="w-3 h-3 rounded-full bg-node-message mr-2" />
            <MessageSquare className="w-4 h-4 text-primary mr-2" />
            <span className="text-sm font-medium text-black">Message</span>
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
            <Textarea
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              className="min-h-[80px] text-sm text-black bg-white border-gray-300"
              placeholder="Enter your message..."
            />
            
            <div className="space-y-2 p-3 bg-muted/30 rounded border">
              <div className="text-xs font-medium text-muted-foreground">Manual Mode Settings:</div>
              <Input
                value={expectedInput}
                onChange={(e) => setExpectedInput(e.target.value)}
                placeholder="Expected user input patterns..."
                className="text-sm"
              />
              <Input
                value={responseOutput}
                onChange={(e) => setResponseOutput(e.target.value)}
                placeholder="Response to give when matched..."
                className="text-sm"
              />
            </div>
            
            <Button size="sm" onClick={handleSave} className="w-full">
              Save
            </Button>
          </div>
        ) : (
          <div className="bg-muted/50 rounded p-3 text-sm text-black">
            {message}
          </div>
        )}
      </div>
      
      <Handle 
        type="source" 
        position={Position.Bottom} 
        id="message-output"
        className="w-3 h-3 bg-primary border-2 border-white"
      />
    </div>
  );
}