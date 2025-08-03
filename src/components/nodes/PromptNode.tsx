import React, { useState, useEffect } from 'react';
import { Handle, Position, NodeProps } from '@xyflow/react';
import { Sparkles, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useMySQLAPI } from '@/hooks/useMySQLAPI';

export default function PromptNode({ data, id }: NodeProps) {
  const [aiSettings, setAiSettings] = useState<any>(null);
  const { callAPI } = useMySQLAPI();

  useEffect(() => {
    loadAISettings();
  }, []);

  const loadAISettings = async () => {
    try {
      const response = await callAPI({
        endpoint: 'mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway',
        method: 'POST',
        data: {
          sql: 'SELECT * FROM ai_settings_nodepath ORDER BY created_at DESC LIMIT 1'
        }
      });

      if (response.success && response.data && response.data.result && response.data.result.length > 0) {
        setAiSettings(response.data.result[0]);
      }
    } catch (error) {
      console.error('Error loading AI settings:', error);
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
              onClick={() => (data?.onDelete as Function)?.(id)}
              className="h-6 w-6 p-0 text-destructive hover:text-destructive"
            >
              <Trash2 className="w-3 h-3" />
            </Button>
          </div>
        </div>
        
        <div className="space-y-2">
          <div className="bg-purple-50 rounded p-3 text-sm">
            <div className="text-xs text-muted-foreground mb-1">System Prompt:</div>
            <div className="text-black">
              {aiSettings?.system_prompt ? (
                aiSettings.system_prompt.length > 100 ? 
                  `${aiSettings.system_prompt.substring(0, 100)}...` : 
                  aiSettings.system_prompt
              ) : (
                'No system prompt configured'
              )}
            </div>
          </div>
          <div className="bg-purple-50 rounded p-3 text-sm">
            <div className="text-xs text-muted-foreground mb-1">Instance Prompt:</div>
            <div className="text-black">
              {aiSettings?.instance_prompt ? (
                aiSettings.instance_prompt.length > 100 ? 
                  `${aiSettings.instance_prompt.substring(0, 100)}...` : 
                  aiSettings.instance_prompt
              ) : (
                'No instance prompt configured'
              )}
            </div>
          </div>
          <div className="text-xs text-muted-foreground italic">
            🤖 AI will use these prompts from AI Settings
          </div>
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