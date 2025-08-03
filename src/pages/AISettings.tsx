import React, { useState, useEffect } from 'react';
import { Card } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Settings, Save, Eye, EyeOff } from 'lucide-react';
import { useMySQLAPI } from '@/hooks/useMySQLAPI';
import { useToast } from '@/hooks/use-toast';

interface AISettingsData {
  id?: string;
  system_prompt?: string;
  closing_prompt?: string;
  instance_prompt?: string;
  open_model?: string;
  open_router_key?: string;
}

export default function AISettings() {
  const [settings, setSettings] = useState<AISettingsData>({
    system_prompt: '',
    closing_prompt: '',
    instance_prompt: '',
    open_model: 'openai/gpt-4.1',
    open_router_key: ''
  });
  const [loading, setLoading] = useState(false);
  const [showApiKey, setShowApiKey] = useState(false);
  const { callAPI } = useMySQLAPI();
  const { toast } = useToast();

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      // First ensure table exists
      await ensureTableExists();
      
      // Use raw SQL to select settings
      const response = await callAPI({
        endpoint: 'mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway',
        method: 'POST',
        data: {
          sql: 'SELECT * FROM ai_settings_nodepath ORDER BY created_at DESC LIMIT 1'
        }
      });

      console.log('Load settings response:', response);
      if (response.success && response.data && response.data.result && response.data.result.length > 0) {
        setSettings(response.data.result[0]);
      }
    } catch (error) {
      console.error('Error loading AI settings:', error);
    }
  };

  const ensureTableExists = async () => {
    try {
      console.log('Creating/checking ai_settings_nodepath table...');
      
      // First create table if it doesn't exist with proper UTF-8 support
      const createTableSQL = `
        CREATE TABLE IF NOT EXISTS ai_settings_nodepath (
          id INT AUTO_INCREMENT PRIMARY KEY,
          system_prompt TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
          closing_prompt TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
          instance_prompt TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
          updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
        ) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci
      `;

      await callAPI({
        endpoint: 'mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway',
        method: 'POST',
        data: {
          sql: createTableSQL
        }
      });

      // Check if new columns exist and add them only if they don't exist
      try {
        // Check for open_model column
        const checkOpenModelSQL = `SHOW COLUMNS FROM ai_settings_nodepath WHERE Field = 'open_model'`;
        const openModelCheck = await callAPI({
          endpoint: 'mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway',
          method: 'POST',
          data: { sql: checkOpenModelSQL }
        });

        // Check for open_router_key column
        const checkOpenRouterSQL = `SHOW COLUMNS FROM ai_settings_nodepath WHERE Field = 'open_router_key'`;
        const openRouterCheck = await callAPI({
          endpoint: 'mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway',
          method: 'POST',
          data: { sql: checkOpenRouterSQL }
        });

        // Add open_model column if it doesn't exist
        if (openModelCheck.success && (!openModelCheck.data?.result || openModelCheck.data.result.length === 0)) {
          try {
            await callAPI({
              endpoint: 'mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway',
              method: 'POST',
              data: { sql: `ALTER TABLE ai_settings_nodepath ADD COLUMN open_model VARCHAR(255) DEFAULT 'openai/gpt-4.1'` }
            });
            console.log('Added open_model column');
          } catch (error) {
            console.log('Error adding open_model column:', error);
          }
        }

        // Add open_router_key column if it doesn't exist
        if (openRouterCheck.success && (!openRouterCheck.data?.result || openRouterCheck.data.result.length === 0)) {
          try {
            await callAPI({
              endpoint: 'mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway',
              method: 'POST',
              data: { sql: `ALTER TABLE ai_settings_nodepath ADD COLUMN open_router_key TEXT` }
            });
            console.log('Added open_router_key column');
          } catch (error) {
            console.log('Error adding open_router_key column:', error);
          }
        }
      } catch (error) {
        console.log('Error checking/adding columns:', error);
      }

      return true;
    } catch (error) {
      console.error('Error ensuring table exists:', error);
      return false;
    }
  };

  const saveSettings = async () => {
    setLoading(true);
    try {
      // Ensure table exists before saving
      const tableCreated = await ensureTableExists();
      if (!tableCreated) {
        throw new Error('Failed to create/verify table');
      }
      
      const settingsData = {
        system_prompt: settings.system_prompt || '',
        closing_prompt: settings.closing_prompt || '',
        instance_prompt: settings.instance_prompt || '',
        open_model: settings.open_model || 'openai/gpt-4.1',
        open_router_key: settings.open_router_key || ''
      };

      console.log('Saving settings with raw SQL approach...');
      console.log('Settings data:', settingsData);

      // Use proper escaping for UTF-8 content including emojis
      const escapeForMySQL = (str: string) => {
        return str
          .replace(/\\/g, '\\\\')
          .replace(/'/g, "\\'")
          .replace(/"/g, '\\"')
          .replace(/\n/g, '\\n')
          .replace(/\r/g, '\\r')
          .replace(/\x00/g, '\\0');
      };

      // Use REPLACE INTO to handle insert/update in one operation
      const insertSQL = `
        REPLACE INTO ai_settings_nodepath (id, system_prompt, closing_prompt, instance_prompt, open_model, open_router_key, updated_at) 
        VALUES (1, '${escapeForMySQL(settingsData.system_prompt)}', '${escapeForMySQL(settingsData.closing_prompt)}', '${escapeForMySQL(settingsData.instance_prompt)}', '${escapeForMySQL(settingsData.open_model)}', '${escapeForMySQL(settingsData.open_router_key)}', NOW())
      `;

      const response = await callAPI({
        endpoint: 'mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway',
        method: 'POST',
        data: {
          sql: insertSQL
        }
      });

      console.log('Save response:', response);
      if (!response.success) {
        throw new Error(response.error || 'Save failed');
      }

      // Update local state
      setSettings(prev => ({ ...prev, id: '1' }));

      toast({
        title: "AI Settings saved",
        description: "Your AI configuration has been updated successfully"
      });
    } catch (error) {
      console.error('Error saving AI settings:', error);
      toast({
        title: "Error saving settings",
        description: `Failed to save AI configuration: ${error.message || error}`,
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

              <div>
                <Label htmlFor="open-model" className="text-base font-semibold text-foreground mb-3 block">
                  Open Model
                </Label>
                <p className="text-sm text-muted-foreground mb-4">
                  AI model to use for processing (hardcoded to Chat GPT 4.1)
                </p>
                <Input
                  id="open-model"
                  value="Chat GPT 4.1 (NEW)"
                  readOnly
                  disabled
                  className="bg-muted"
                />
                <p className="text-xs text-muted-foreground mt-2">
                  Model value: openai/gpt-4.1
                </p>
              </div>

              <div>
                <Label htmlFor="open-router-key" className="text-base font-semibold text-foreground mb-3 block">
                  OpenRouter API Key
                </Label>
                <p className="text-sm text-muted-foreground mb-4">
                  Your OpenRouter API key for AI model access
                </p>
                <div className="relative">
                  <Input
                    id="open-router-key"
                    type={showApiKey ? "text" : "password"}
                    placeholder="Enter your OpenRouter API key..."
                    value={settings.open_router_key || ''}
                    onChange={(e) => handleInputChange('open_router_key', e.target.value)}
                    className="pr-10"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="absolute right-2 top-1/2 -translate-y-1/2 h-6 w-6 p-0 hover:bg-transparent"
                    onClick={() => setShowApiKey(!showApiKey)}
                  >
                    {showApiKey ? (
                      <EyeOff className="h-4 w-4 text-muted-foreground" />
                    ) : (
                      <Eye className="h-4 w-4 text-muted-foreground" />
                    )}
                  </Button>
                </div>
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