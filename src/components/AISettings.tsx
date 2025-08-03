import React, { useState, useEffect } from 'react';
import { Card } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Button } from '@/components/ui/button';
import { Settings, Save } from 'lucide-react';
import { supabase } from '@/integrations/supabase/client';
import { useToast } from '@/hooks/use-toast';

interface AISettingsData {
  id?: string;
  system_prompt?: string;
  closing_prompt?: string;
  instance_prompt?: string;
}

export default function AISettings() {
  const [settings, setSettings] = useState<AISettingsData>({
    system_prompt: '',
    closing_prompt: '',
    instance_prompt: ''
  });
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      const { data, error } = await supabase
        .from('ai_settings_nodepath')
        .select('*')
        .limit(1)
        .single();

      if (error && error.code !== 'PGRST116') { // PGRST116 is "no rows returned"
        console.error('Error loading AI settings:', error);
        return;
      }

      if (data) {
        setSettings(data);
      }
    } catch (error) {
      console.error('Error loading AI settings:', error);
    }
  };

  const saveSettings = async () => {
    setLoading(true);
    try {
      const settingsData = {
        system_prompt: settings.system_prompt || '',
        closing_prompt: settings.closing_prompt || '',
        instance_prompt: settings.instance_prompt || '',
        updated_at: new Date().toISOString()
      };

      if (settings.id) {
        // Update existing settings
        const { error } = await supabase
          .from('ai_settings_nodepath')
          .update(settingsData)
          .eq('id', settings.id);

        if (error) throw error;
      } else {
        // Create new settings
        const { data, error } = await supabase
          .from('ai_settings_nodepath')
          .insert([settingsData])
          .select()
          .single();

        if (error) throw error;
        if (data) {
          setSettings(prev => ({ ...prev, id: data.id }));
        }
      }

      toast({
        title: "AI Settings saved",
        description: "Your AI configuration has been updated successfully"
      });
    } catch (error) {
      console.error('Error saving AI settings:', error);
      toast({
        title: "Error saving settings",
        description: "Failed to save AI configuration",
        variant: "destructive"
      });
    } finally {
      setLoading(false);
    }
  };

  const handleInputChange = (field: keyof AISettingsData, value: string) => {
    setSettings(prev => ({ ...prev, [field]: value }));
  };

  return (
    <Card className="w-full bg-card border-border">
      <div className="p-6">
        <div className="flex items-center gap-2 mb-6">
          <Settings className="w-5 h-5" />
          <h3 className="text-lg font-semibold text-foreground">AI Settings</h3>
        </div>
        
        <div className="space-y-6">
          <div>
            <Label htmlFor="system-prompt" className="text-sm font-medium text-foreground mb-2 block">
              System Prompt
            </Label>
            <Textarea
              id="system-prompt"
              placeholder="Define AI's personality or behavior (e.g., friendly, professional, humorous)..."
              value={settings.system_prompt || ''}
              onChange={(e) => handleInputChange('system_prompt', e.target.value)}
              className="min-h-[80px] resize-none"
            />
          </div>

          <div>
            <Label htmlFor="closing-prompt" className="text-sm font-medium text-foreground mb-2 block">
              Closing Prompt
            </Label>
            <Textarea
              id="closing-prompt"
              placeholder="Used when ending a conversation (e.g., thank you, CTA, etc.)..."
              value={settings.closing_prompt || ''}
              onChange={(e) => handleInputChange('closing_prompt', e.target.value)}
              className="min-h-[80px] resize-none"
            />
          </div>

          <div>
            <Label htmlFor="instance-prompt" className="text-sm font-medium text-foreground mb-2 block">
              Instance Prompt
            </Label>
            <Textarea
              id="instance-prompt"
              placeholder="Define how AI should handle multiple instances (e.g., simultaneous chats or context switch rules)..."
              value={settings.instance_prompt || ''}
              onChange={(e) => handleInputChange('instance_prompt', e.target.value)}
              className="min-h-[80px] resize-none"
            />
          </div>

          <Button 
            onClick={saveSettings}
            disabled={loading}
            variant="default"
            className="w-full"
          >
            <Save className="w-4 h-4 mr-2" />
            {loading ? 'Saving...' : 'Save AI Settings'}
          </Button>
        </div>
      </div>
    </Card>
  );
}