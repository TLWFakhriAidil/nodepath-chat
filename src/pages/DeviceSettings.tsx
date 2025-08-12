import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Textarea } from '@/components/ui/textarea';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { toast } from 'sonner';
import { Smartphone, Settings, Save, X, Link, Copy, Plus, Edit, Trash2, Eye } from 'lucide-react';

interface DeviceSettings {
  id: string;
  device_id: string;
  api_key_option: string;
  webhook_id: string;
  provider: string;
  phone_number: string;
  api_key: string;
  id_device: string;
  id_erp: string;
  id_admin: string;
}

// DeviceForm component moved outside to prevent re-creation on every render
const DeviceForm: React.FC<{
  settings: DeviceSettings;
  handleInputChange: (field: keyof DeviceSettings, value: string) => void;
  generateDeviceId: () => void;
  generateWebhookId: () => void;
  handleSave: () => void;
  handleClose: () => void;
  isSaving: boolean;
  apiKeyOptions: Array<{value: string; label: string}>;
  providerOptions: Array<{value: string; label: string}>;
}> = ({ settings, handleInputChange, generateDeviceId, generateWebhookId, handleSave, handleClose, isSaving, apiKeyOptions, providerOptions }) => (
  <div className="space-y-6">
    {/* Device ID Section */}
    <div className="space-y-4">
      <div>
        <Label className="text-slate-700 dark:text-slate-300 font-medium">Device ID (VIEW ONLY)</Label>
        <div className="flex gap-2 mt-1">
          <Input
            value={settings.device_id}
            placeholder="Device ID will appear here"
            className="bg-slate-50 dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-900 dark:text-slate-100"
            readOnly
          />
          <Button
            onClick={generateDeviceId}
            className="bg-gradient-to-r from-blue-500 to-purple-600 text-white hover:from-blue-600 hover:to-purple-700"
          >
            GENERATE DEVICE
          </Button>
        </div>
      </div>

      <div>
        <Label className="text-slate-700 dark:text-slate-300 font-medium">Webhook ID</Label>
        <div className="flex gap-2 mt-1">
          <Input
            value={settings.webhook_id}
            onChange={(e) => handleInputChange('webhook_id', e.target.value)}
            placeholder="https://chatbot.growweb.com/chatgpt/SCVTC-S2/FGcaTDgH"
            className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-900 dark:text-slate-100"
          />
          <Button
            onClick={generateWebhookId}
            className="bg-gradient-to-r from-green-500 to-emerald-600 text-white hover:from-green-600 hover:to-emerald-700 flex items-center gap-2"
          >
            <Link className="h-4 w-4" />
            GENERATE WEBHOOK
          </Button>
        </div>
      </div>
    </div>

    {/* API Key Options */}
    <div>
      <Label className="text-slate-700 dark:text-slate-300 font-medium mb-3 block">API Key Option</Label>
      <RadioGroup
        value={settings.api_key_option}
        onValueChange={(value) => handleInputChange('api_key_option', value)}
        className="space-y-2"
      >
        {apiKeyOptions.map((option) => (
          <div key={option.value} className="flex items-center space-x-2">
            <RadioGroupItem
              value={option.value}
              id={option.value}
              className="border-slate-300 dark:border-slate-600"
            />
            <Label htmlFor={option.value} className="text-slate-700 dark:text-slate-300">
              {option.label}
            </Label>
          </div>
        ))}
      </RadioGroup>
    </div>

    {/* Provider Options */}
    <div>
      <Label className="text-slate-700 dark:text-slate-300 font-medium mb-3 block">Provider</Label>
      <RadioGroup
        value={settings.provider}
        onValueChange={(value) => handleInputChange('provider', value)}
        className="space-y-2"
      >
        {providerOptions.map((option) => (
          <div key={option.value} className="flex items-center space-x-2">
            <RadioGroupItem
              value={option.value}
              id={option.value}
              className="border-slate-300 dark:border-slate-600"
            />
            <Label htmlFor={option.value} className="text-slate-700 dark:text-slate-300">
              {option.label}
            </Label>
          </div>
        ))}
      </RadioGroup>
    </div>

    {/* Phone Number */}
    <div>
      <Label className="text-slate-700 dark:text-slate-300 font-medium">Phone Number</Label>
      <Input
        value={settings.phone_number}
        onChange={(e) => {
          // Only allow numbers, spaces, hyphens, parentheses, and plus sign
          const value = e.target.value.replace(/[^0-9\s\-\(\)\+]/g, '');
          handleInputChange('phone_number', value);
        }}
        placeholder="Enter phone number (numbers only)"
        className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-900 dark:text-slate-100 mt-1"
        type="tel"
        pattern="[0-9\s\-\(\)\+]*"
      />
    </div>

    {/* API Key */}
    <div>
      <Label className="text-slate-700 dark:text-slate-300 font-medium">API Key https://openrouter.ai</Label>
      <Textarea
        value={settings.api_key}
        onChange={(e) => handleInputChange('api_key', e.target.value)}
        placeholder="sk-or-v1-Sa726e885f027c95ee8142f0ae3ee6af6ff1bf0cd6df"
        className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-900 dark:text-slate-100 mt-1 min-h-[80px]"
      />
    </div>

    {/* Required Input Fields */}
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div>
        <Label className="text-slate-700 dark:text-slate-300 font-medium">ID Device</Label>
        <Input
          value={settings.id_device}
          onChange={(e) => handleInputChange('id_device', e.target.value)}
          placeholder="Enter ID Device"
          className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-900 dark:text-slate-100 mt-1"
          required
        />
      </div>
      <div>
        <Label className="text-slate-700 dark:text-slate-300 font-medium">ID ERP</Label>
        <Input
          value={settings.id_erp}
          onChange={(e) => handleInputChange('id_erp', e.target.value)}
          placeholder="Enter ID ERP"
          className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-900 dark:text-slate-100 mt-1"
          required
        />
      </div>
      <div>
        <Label className="text-slate-700 dark:text-slate-300 font-medium">ID Admin</Label>
        <Input
          value={settings.id_admin}
          onChange={(e) => handleInputChange('id_admin', e.target.value)}
          placeholder="Enter ID Admin"
          className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-900 dark:text-slate-100 mt-1"
          required
        />
      </div>
    </div>

    {/* Action Buttons */}
    <div className="flex gap-3 pt-4 border-t border-slate-200 dark:border-slate-700">
      <Button
        onClick={handleClose}
        variant="outline"
        className="border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800"
      >
        Cancel
      </Button>
      <Button
        onClick={handleSave}
        disabled={isSaving || !settings.id_device || !settings.id_erp || !settings.id_admin}
        className="bg-gradient-to-r from-blue-500 to-purple-600 text-white hover:from-blue-600 hover:to-purple-700 flex items-center gap-2"
      >
        {isSaving ? (
          <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
        ) : (
          <Save className="h-4 w-4" />
        )}
        {isSaving ? 'Saving...' : 'Save Device'}
      </Button>
    </div>
  </div>
);

const DeviceSettings: React.FC = () => {
  const [devices, setDevices] = useState<DeviceSettings[]>([]);
  const [settings, setSettings] = useState<DeviceSettings>({
    id: '',
    device_id: '',
    api_key_option: 'chat_gpt_4_1_new',
    webhook_id: '',
    provider: 'wablas',
    phone_number: '',
    api_key: '',
    id_device: '',
    id_erp: '',
    id_admin: ''
  });
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingDevice, setEditingDevice] = useState<DeviceSettings | null>(null);

  const apiKeyOptions = [
    { value: 'chat_gpt_so', label: 'Chat GPT So' },
    { value: 'chat_gpt_5_mini', label: 'Chat GPT 5 Mini' },
    { value: 'chat_gpt_4o', label: 'Chat GPT 4o' },
    { value: 'chat_gpt_4_1_new', label: 'Chat GPT 4.1 (NEW)' },
    { value: 'gemini_pro_25', label: 'GEMINI PRO 2.5' },
    { value: 'gemini_pro_15', label: 'GEMINI PRO 1.5' }
  ];

  const providerOptions = [
    { value: 'whacenter', label: 'Whacenter' },
    { value: 'wablas', label: 'Wablas' },
    { value: 'rvsb_wasap', label: 'RVSB WASAP' }
  ];

  useEffect(() => {
    loadDeviceSettings();
  }, []);

  const loadDeviceSettings = async () => {
    setIsLoading(true);
    try {
      const response = await fetch('/api/device-settings');
      if (response.ok) {
        const data = await response.json();
        // Ensure data is an array, handle different response formats
        if (Array.isArray(data)) {
          setDevices(data);
        } else if (data && Array.isArray(data.data)) {
          setDevices(data.data);
        } else if (data && data.success && Array.isArray(data.data)) {
          setDevices(data.data);
        } else {
          console.warn('Unexpected API response format:', data);
          setDevices([]);
        }
      } else {
        console.error('API response not ok:', response.status, response.statusText);
        setDevices([]);
      }
    } catch (error) {
      console.error('Error loading device settings:', error);
      toast.error('Failed to load device settings');
      setDevices([]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSave = async () => {
    setIsSaving(true);
    try {
      const method = settings.id ? 'PUT' : 'POST';
      const url = settings.id ? `/api/device-settings/${settings.id}` : '/api/device-settings';
      
      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(settings),
      });

      if (response.ok) {
        const savedSettings = await response.json();
        toast.success('Device settings saved successfully!');
        setIsModalOpen(false);
        resetForm();
        loadDeviceSettings(); // Reload the table
      } else {
        throw new Error('Failed to save settings');
      }
    } catch (error) {
      console.error('Error saving device settings:', error);
      toast.error('Failed to save device settings');
    } finally {
      setIsSaving(false);
    }
  };

  const handleInputChange = (field: keyof DeviceSettings, value: string) => {
    setSettings(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleClose = () => {
    setIsModalOpen(false);
    resetForm();
  };

  const resetForm = () => {
    setSettings({
      id: '',
      device_id: '',
      api_key_option: 'chat_gpt_4_1_new',
      webhook_id: '',
      provider: 'wablas',
      phone_number: '',
      api_key: '',
      id_device: '',
      id_erp: '',
      id_admin: ''
    });
    setEditingDevice(null);
  };

  const handleNewDevice = () => {
    resetForm();
    setIsModalOpen(true);
  };

  const handleEditDevice = (device: DeviceSettings) => {
    setSettings(device);
    setEditingDevice(device);
    setIsModalOpen(true);
  };

  const handleDeleteDevice = async (deviceId: string) => {
    if (!confirm('Are you sure you want to delete this device?')) return;
    
    try {
      const response = await fetch(`/api/device-settings/${deviceId}`, {
        method: 'DELETE',
      });
      
      if (response.ok) {
        toast.success('Device deleted successfully!');
        loadDeviceSettings();
      } else {
        throw new Error('Failed to delete device');
      }
    } catch (error) {
      console.error('Error deleting device:', error);
      toast.error('Failed to delete device');
    }
  };

  const generateDeviceId = async () => {
    // Validation: Check if phone number and id_device are provided
    if (!settings.phone_number.trim()) {
      toast.error('Please enter a phone number before generating device');
      return;
    }
    
    if (!settings.id_device.trim()) {
      toast.error('Please enter ID Device before generating device');
      return;
    }

    if (!settings.phone_number.trim() && !settings.id_device.trim()) {
      toast.error('Please enter both phone number and ID Device before generating device');
      return;
    }

    setIsSaving(true);
    toast.info('Generating device... Please wait');

    try {
      // Get Railway deployment URL from window.location or use environment
      const baseUrl = window.location.origin;
      const webhookUrl = `${baseUrl}/api/webhook/${settings.id_device}`;
      
      let apiResponse;
      
      if (settings.provider === 'whacenter') {
        // Whacenter API integration
        const whacenterData = {
          device_name: settings.id_device,
          webhook_url: webhookUrl
        };
        
        apiResponse = await fetch('/api/device-settings/generate-whacenter', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            ...settings,
            webhook_url: webhookUrl,
            device_data: whacenterData
          })
        });
      } else if (settings.provider === 'wablas') {
        // Wablas API integration
        const wablasData = {
          device_name: settings.id_device,
          webhook_url: webhookUrl
        };
        
        apiResponse = await fetch('/api/device-settings/generate-wablas', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            ...settings,
            webhook_url: webhookUrl,
            device_data: wablasData
          })
        });
      } else {
        // Fallback: Generate local device ID
        const deviceId = `C${Math.floor(Math.random() * 10000).toString().padStart(4, '0')}H`;
        handleInputChange('device_id', deviceId);
        handleInputChange('webhook_id', webhookUrl);
        toast.success('Device ID generated successfully!');
        setIsSaving(false);
        return;
      }

      if (apiResponse && apiResponse.ok) {
        const result = await apiResponse.json();
        
        if (result.success) {
          // Update form with generated data
          if (result.data.device_id) {
            handleInputChange('device_id', result.data.device_id);
          }
          if (result.data.webhook_url) {
            handleInputChange('webhook_id', result.data.webhook_url);
          }
          if (result.data.api_key) {
            handleInputChange('api_key', result.data.api_key);
          }
          
          toast.success(`Device generated successfully via ${settings.provider}!`);
        } else {
          throw new Error(result.message || 'Failed to generate device');
        }
      } else {
        throw new Error('Failed to communicate with device provider');
      }
    } catch (error) {
      console.error('Error generating device:', error);
      toast.error(`Failed to generate device: ${error.message}`);
      
      // Fallback: Generate local device ID
      const deviceId = `C${Math.floor(Math.random() * 10000).toString().padStart(4, '0')}H`;
      handleInputChange('device_id', deviceId);
      toast.info('Generated local device ID as fallback');
    } finally {
      setIsSaving(false);
    }
  };

  const generateWebhookId = () => {
    if (!settings.id_device.trim()) {
      toast.error('Please enter ID Device first before generating webhook');
      return;
    }
    
    // Get Railway deployment URL from window.location
    const baseUrl = window.location.origin;
    const webhookId = `${baseUrl}/api/webhook/${settings.id_device}`;
    handleInputChange('webhook_id', webhookId);
    toast.success('Webhook ID generated successfully!');
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }



  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white mb-2">
            Device Settings
          </h1>
          <p className="text-slate-600 dark:text-slate-400">
            Manage all your device configurations and webhook integrations
          </p>
        </div>
        <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
          <DialogTrigger asChild>
            <Button 
              onClick={handleNewDevice}
              className="bg-gradient-to-r from-blue-500 to-purple-600 text-white hover:from-blue-600 hover:to-purple-700 flex items-center gap-2"
            >
              <Plus className="h-4 w-4" />
              New Device
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                <Smartphone className="h-5 w-5" />
                {editingDevice ? 'Edit Device' : 'Add New Device'}
              </DialogTitle>
            </DialogHeader>
            <DeviceForm 
              settings={settings}
              handleInputChange={handleInputChange}
              generateDeviceId={generateDeviceId}
              generateWebhookId={generateWebhookId}
              handleSave={handleSave}
              handleClose={handleClose}
              isSaving={isSaving}
              apiKeyOptions={apiKeyOptions}
              providerOptions={providerOptions}
            />
          </DialogContent>
        </Dialog>
      </div>

      {/* Devices Table */}
      <Card className="border-0 shadow-xl">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Settings className="h-5 w-5" />
            All Devices
          </CardTitle>
        </CardHeader>
        <CardContent>
          {devices.length === 0 ? (
             <div className="text-center py-8">
               <Smartphone className="h-12 w-12 text-slate-400 mx-auto mb-4" />
               <h3 className="text-lg font-medium text-slate-900 dark:text-white mb-2">No devices found</h3>
               <p className="text-slate-600 dark:text-slate-400">Get started by adding your first device configuration using the "New Device" button above.</p>
             </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Device ID</TableHead>
                  <TableHead>Provider</TableHead>
                  <TableHead>Phone Number</TableHead>
                  <TableHead>API Key Option</TableHead>
                  <TableHead>ID Device</TableHead>
                  <TableHead>ID ERP</TableHead>
                  <TableHead>ID Admin</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {devices.map((device) => (
                  <TableRow key={device.id}>
                    <TableCell className="font-medium">{device.device_id || 'Not generated'}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{device.provider}</Badge>
                    </TableCell>
                    <TableCell>{device.phone_number || 'Not set'}</TableCell>
                    <TableCell>{apiKeyOptions.find(opt => opt.value === device.api_key_option)?.label}</TableCell>
                    <TableCell>{device.id_device}</TableCell>
                    <TableCell>{device.id_erp}</TableCell>
                    <TableCell>{device.id_admin}</TableCell>
                    <TableCell>
                      <Badge variant={device.device_id ? 'default' : 'secondary'}>
                        {device.device_id ? 'Active' : 'Pending'}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleEditDevice(device)}
                          className="h-8 w-8 p-0"
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDeleteDevice(device.id)}
                          className="h-8 w-8 p-0 text-red-600 hover:text-red-700 hover:bg-red-50"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
};



export default DeviceSettings;