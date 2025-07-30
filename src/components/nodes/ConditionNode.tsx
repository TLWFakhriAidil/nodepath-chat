import React, { useState } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { GitBranch, Edit3, Trash2, Plus, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ConditionRule } from '@/types/chatbot';

export default function ConditionNode({ data, id }: NodeProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [conditions, setConditions] = useState<ConditionRule[]>(
    (data?.conditions as ConditionRule[]) || [
      { id: '1', type: 'contains', value: 'yes', label: 'Yes', nextNodeId: '' },
      { id: '2', type: 'default', label: 'Default', nextNodeId: '' }
    ]
  );

  const handleSave = () => {
    setIsEditing(false);
    if (data?.onUpdate) {
      (data.onUpdate as Function)(id, { conditions });
    }
  };

  const addCondition = () => {
    const newCondition: ConditionRule = {
      id: Date.now().toString(),
      type: 'contains',
      value: '',
      label: 'New condition',
      nextNodeId: ''
    };
    setConditions([...conditions, newCondition]);
  };

  const updateCondition = (conditionId: string, updates: Partial<ConditionRule>) => {
    setConditions(conditions.map(c => 
      c.id === conditionId ? { ...c, ...updates } : c
    ));
  };

  const removeCondition = (conditionId: string) => {
    setConditions(conditions.filter(c => c.id !== conditionId));
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
            {conditions.map((condition) => (
              <div key={condition.id} className="space-y-2 p-3 border rounded">
                <div className="flex justify-between items-center">
                  <Input
                    value={condition.label}
                    onChange={(e) => updateCondition(condition.id, { label: e.target.value })}
                    placeholder="Condition label"
                    className="text-sm"
                  />
                  {conditions.length > 1 && (
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => removeCondition(condition.id)}
                      className="h-6 w-6 p-0 ml-2"
                    >
                      <X className="w-3 h-3" />
                    </Button>
                  )}
                </div>
                <Select
                  value={condition.type}
                  onValueChange={(value) => updateCondition(condition.id, { type: value as any })}
                >
                  <SelectTrigger className="text-sm">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="equals">Equals</SelectItem>
                    <SelectItem value="contains">Contains</SelectItem>
                    <SelectItem value="default">Default (fallback)</SelectItem>
                  </SelectContent>
                </Select>
                {condition.type !== 'default' && (
                  <Input
                    value={condition.value || ''}
                    onChange={(e) => updateCondition(condition.id, { value: e.target.value })}
                    placeholder="Value to match"
                    className="text-sm"
                  />
                )}
              </div>
            ))}
            <Button size="sm" onClick={addCondition} variant="outline" className="w-full">
              <Plus className="w-3 h-3 mr-2" />
              Add Condition
            </Button>
            <Button size="sm" onClick={handleSave} className="w-full">
              Save
            </Button>
          </div>
        ) : (
          <div className="bg-muted/50 rounded p-3 text-sm text-foreground space-y-1">
            {conditions.map((condition) => (
              <div key={condition.id} className="flex justify-between">
                <span className="font-medium">{condition.label}:</span>
                <span className="text-muted-foreground">
                  {condition.type === 'default' ? 'Default' : `${condition.type} "${condition.value}"`}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
      
      {/* Condition outputs */}
      <div className="flex justify-between px-4 pb-2 text-xs font-medium">
        {conditions.slice(0, 2).map((condition, index) => (
          <div key={condition.id} className={index === 0 ? "text-green-500" : "text-red-500"}>
            {condition.label}
          </div>
        ))}
      </div>
      
      {/* Dynamic handles based on conditions */}
      {conditions.slice(0, 2).map((condition, index) => (
        <Handle 
          key={condition.id}
          id={condition.id}
          type="source" 
          position={Position.Bottom} 
          style={{ left: `${25 + (index * 50)}%` }}
          className={`w-3 h-3 border-2 border-white ${index === 0 ? 'bg-green-500' : 'bg-red-500'}`}
        />
      ))}
    </div>
  );
}