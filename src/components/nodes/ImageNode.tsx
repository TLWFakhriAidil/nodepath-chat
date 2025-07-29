import React, { useState } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Image, Edit3, Upload, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function ImageNode({ data, id }: NodeProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [imageUrl, setImageUrl] = useState((data?.imageUrl as string) || '');
  const [caption, setCaption] = useState((data?.caption as string) || 'Image caption...');

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
            <div className="w-3 h-3 rounded-full bg-blue-500 mr-2" />
            <Image className="w-4 h-4 text-blue-500 mr-2" />
            <span className="text-sm font-medium text-foreground">Send Image</span>
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
            <div>
              <label className="text-xs text-muted-foreground mb-1 block">Image URL</label>
              <Input
                value={imageUrl}
                onChange={(e) => setImageUrl(e.target.value)}
                className="text-sm"
                placeholder="https://example.com/image.jpg"
              />
            </div>
            <div>
              <label className="text-xs text-muted-foreground mb-1 block">Caption</label>
              <Input
                value={caption}
                onChange={(e) => setCaption(e.target.value)}
                className="text-sm"
                placeholder="Image caption..."
              />
            </div>
            <Button size="sm" onClick={handleSave} className="w-full">
              Save
            </Button>
          </div>
        ) : (
          <div className="space-y-2">
            {imageUrl && (
              <div className="bg-muted/50 rounded p-2 text-center">
                <Upload className="w-8 h-8 mx-auto text-muted-foreground mb-1" />
                <div className="text-xs text-muted-foreground truncate">{imageUrl}</div>
              </div>
            )}
            <div className="bg-muted/50 rounded p-3 text-sm text-foreground">
              {caption}
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