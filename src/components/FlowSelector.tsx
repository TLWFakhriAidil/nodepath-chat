import React from 'react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { ChatbotFlow } from '@/types/chatbot';

interface FlowSelectorProps {
  flows: ChatbotFlow[];
  selectedFlowId: string | null;
  onFlowSelect: (flowId: string) => void;
}

export default function FlowSelector({ flows, selectedFlowId, onFlowSelect }: FlowSelectorProps) {
  return (
    <Select value={selectedFlowId || ''} onValueChange={onFlowSelect}>
      <SelectTrigger className="w-full">
        <SelectValue placeholder="Select a flow to preview" />
      </SelectTrigger>
      <SelectContent>
        {flows.length === 0 ? (
          <SelectItem value="" disabled>
            No flows available
          </SelectItem>
        ) : (
          flows.map((flow) => (
            <SelectItem key={flow.id} value={flow.id}>
              <div className="flex flex-col items-start">
                <span className="font-medium">{flow.name}</span>
                <span className="text-xs text-muted-foreground">
                  {flow.nodes.length} nodes • {new Date(flow.updatedAt).toLocaleDateString()}
                </span>
              </div>
            </SelectItem>
          ))
        )}
      </SelectContent>
    </Select>
  );
}