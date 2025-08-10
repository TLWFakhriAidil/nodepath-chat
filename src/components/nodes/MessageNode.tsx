import React, { useState } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { MessageSquare, Edit3, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Input } from '@/components/ui/input';

export default function MessageNode({ data, id }: NodeProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [message, setMessage] = useState((data?.message as string) || 'Enter your message here...');

  const handleSave = () => {
    setIsEditing(false);
    // Update the node data through the parent
    if (data?.onUpdate) {
      (data.onUpdate as Function)(id, { message });
    }
  };

  return (
    <div className="bg-card/95 backdrop-blur-xl rounded-lg shadow-node border border-border/50 min-w-[250px] max-w-[350px] futuristic-border glow-on-hover transition-all duration-300">
      <Handle 
        type="target" 
        position={Position.Top} 
        id="message-input"
        className="w-3 h-3 bg-primary border-2 border-background shadow-lg"
      />
      
      <div className="p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center">
            <div className="w-3 h-3 rounded-full bg-gradient-to-r from-blue-500 to-blue-600 mr-2 pulse-glow" />
            <MessageSquare className="w-4 h-4 text-primary mr-2" />
            <span className="text-sm font-medium text-foreground holographic-text">Message</span>
          </div>
          <div className="flex items-center gap-1">
            <Button
              size="sm"
              variant="ghost"
              onClick={() => setIsEditing(!isEditing)}
              className="h-6 w-6 p-0 hover:bg-primary/20 hover:text-primary transition-colors"
            >
              <Edit3 className="w-3 h-3" />
            </Button>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => (data?.onDelete as Function)?.(id)}
              className="h-6 w-6 p-0 text-destructive hover:text-destructive hover:bg-destructive/20 transition-colors"
            >
              <Trash2 className="w-3 h-3" />
            </Button>
          </div>
        </div>
        
        {isEditing ? (
          <div className="space-y-2">
            <Textarea
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              className="min-h-[80px] text-sm text-foreground bg-input/50 backdrop-blur border-border/50 focus:border-primary focus:ring-primary/20 transition-all duration-300"
              placeholder="Enter your message..."
            />
            <Button 
              size="sm" 
              onClick={handleSave} 
              className="w-full bg-gradient-to-r from-primary to-blue-600 hover:from-blue-600 hover:to-primary text-white glow-on-hover transition-all duration-300"
            >
              Save
            </Button>
          </div>
        ) : (
          <div className="bg-muted/30 backdrop-blur rounded p-3 text-sm text-foreground border border-border/30">
            {message}
          </div>
        )}
      </div>
      
      <Handle 
        type="source" 
        position={Position.Bottom} 
        id="message-output"
        className="w-3 h-3 bg-primary border-2 border-background shadow-lg"
      />
    </div>
  );
}