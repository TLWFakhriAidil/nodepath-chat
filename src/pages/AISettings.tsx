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
    <div className="container mx-auto py-8 px-6">
      <div className="max-w-4xl mx-auto">
        <div className="flex items-center gap-3 mb-8">
          <Settings className="w-8 h-8 text-primary" />
          <div>
            <h1 className="text-3xl font-bold text-foreground">AI Settings</h1>
            <p className="text-muted-foreground">Configure how the AI behaves across all flows</p>
          </div>
        </div>

        <Card className="bg-card border-border">
          <div className="p-8">
            <div className="space-y-8">
              <div>
                <Label htmlFor="system-prompt" className="text-base font-semibold text-foreground mb-3 block">
                  System Prompt
                </Label>
                <p className="text-sm text-muted-foreground mb-4">
                  Define AI's personality or behavior (e.g., friendly, professional, humorous)
                </p>
                <Textarea
                  id="system-prompt"
                  placeholder="Enter system prompt that defines the AI's overall behavior and personality..."
                  value={settings.system_prompt || ''}
                  onChange={(e) => handleInputChange('system_prompt', e.target.value)}
                  className="min-h-[120px] resize-none"
                />
              </div>

              <div>
                <Label htmlFor="closing-prompt" className="text-base font-semibold text-foreground mb-3 block">
                  Closing Prompt
                </Label>
                <p className="text-sm text-muted-foreground mb-4">
                  Used when ending a conversation (e.g., thank you message, call-to-action, contact information)
                </p>
                <Textarea
                  id="closing-prompt"
                  placeholder="Enter closing prompt used when conversations end..."
                  value={settings.closing_prompt || ''}
                  onChange={(e) => handleInputChange('closing_prompt', e.target.value)}
                  className="min-h-[120px] resize-none"
                />
              </div>

              <div>
                <Label htmlFor="instance-prompt" className="text-base font-semibold text-foreground mb-3 block">
                  Instance Prompt
                </Label>
                <p className="text-sm text-muted-foreground mb-4">
                  Define how AI should handle multiple instances (e.g., simultaneous chats or context switch rules)
                </p>
                <Textarea
                  id="instance-prompt"
                  placeholder="Enter instance prompt for handling multiple conversations or context switching..."
                  value={settings.instance_prompt || ''}
                  onChange={(e) => handleInputChange('instance_prompt', e.target.value)}
                  className="min-h-[120px] resize-none"
                />
              </div>

              <div className="pt-6">
                <Button 
                  onClick={saveSettings}
                  disabled={loading}
                  variant="default"
                  size="lg"
                  className="px-8"
                >
                  <Save className="w-5 h-5 mr-2" />
                  {loading ? 'Saving Settings...' : 'Save AI Settings'}
                </Button>
              </div>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}