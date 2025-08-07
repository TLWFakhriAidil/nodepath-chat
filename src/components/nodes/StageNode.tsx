import React, { useState } from 'react';
import { Handle, Position } from '@xyflow/react';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { X, Flag } from 'lucide-react';

interface StageNodeProps {
  data: {
    label: string;
    stageName?: string;
    onDelete: (id: string) => void;
    onUpdate: (id: string, data: any) => void;
  };
  id: string;
}

export default function StageNode({ data, id }: StageNodeProps) {
  const [stageName, setStageName] = useState(data.stageName || '');

  const handleStageNameChange = (value: string) => {
    setStageName(value);
    data.onUpdate(id, { stageName: value });
  };

  return (
    <Card className="min-w-[200px] bg-card border-border">
      <div className="flex items-center justify-between p-3 border-b border-border">
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded-full bg-orange-500" />
          <Flag className="w-4 h-4 text-foreground" />
          <span className="text-sm font-medium text-foreground">Stage</span>
        </div>
        <Button
          size="sm"
          variant="ghost"
          onClick={() => data.onDelete(id)}
          className="h-6 w-6 p-0 hover:bg-destructive hover:text-destructive-foreground"
        >
          <X className="w-3 h-3" />
        </Button>
      </div>
      
      <div className="p-3 space-y-3">
        <div className="space-y-2">
          <Label htmlFor={`stage-name-${id}`} className="text-xs text-muted-foreground">
            Stage Name (varchar)
          </Label>
          <Input
            id={`stage-name-${id}`}
            value={stageName}
            onChange={(e) => handleStageNameChange(e.target.value)}
            placeholder="Enter stage name..."
            className="text-sm"
          />
        </div>
      </div>

      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 bg-border border-2 border-background"
      />
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-3 h-3 bg-border border-2 border-background"
      />
    </Card>
  );
}