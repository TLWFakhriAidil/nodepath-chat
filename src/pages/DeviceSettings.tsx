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
  setSettings: React.Dispatch<React.SetStateAction<DeviceSettings>>;
}> = ({ settings, handleInputChange, generateDeviceId, generateWebhookId, handleSave, handleClose, isSaving, apiKeyOptions, providerOptions, setSettings }) => (
  <div className="space-y-6">
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
          <div className="flex gap-2">
            <Button
              onClick={() => {
                generateDeviceId();
              }}
              disabled={isSaving}
              className="bg-gradient-to-r from-blue-500 to-purple-600 text-white hover:from-blue-600 hover:to-purple-700 disabled:opacity-50 disabled:cursor-not-allowed flex-1"
            >
              {isSaving ? 'GENERATING...' : 'GENERATE DEVICE'}
            </Button>
          </div>
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

    <div>
      <Label className="text-slate-700 dark:text-slate-300 font-medium">Phone Number</Label>
      <Input
        value={settings.phone_number}
        onChange={(e) => {
          const value = e.target.value.replace(/[^0-9\s\-\(\)\+]/g, '');
          handleInputChange('phone_number', value);
        }}
        placeholder="Enter phone number (numbers only)"
        className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-900 dark:text-slate-100 mt-1"
        type="tel"
        pattern="[0-9\s\-\(\)\+]*"
      />
    </div>

    <div>
      <Label className="text-slate-700 dark:text-slate-300 font-medium">API Key https://openrouter.ai</Label>
      <Textarea
        value={settings.api_key}
        onChange={(e) => handleInputChange('api_key', e.target.value)}
        placeholder="sk-or-v1-Sa726e885f027c95ee8142f0ae3ee6af6ff1bf0cd6df"
        className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-900 dark:text-slate-100 mt-1 min-h-[80px]"
      />
    </div>

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
          <div className="animate-spin rounded-full h-4 w-4"></div>
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
  const [loading, setLoading] = useState(true);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingDevice, setEditingDevice] = useState<DeviceSettings | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [currentSettings, setCurrentSettings] = useState<DeviceSettings>({
    id: '',
    device_id: '',
    api_key_option: 'default',
    webhook_id: '',
    provider: 'whatsapp',
    phone_number: '',
    api_key: '',
    id_device: '',
    id_erp: '',
    id_admin: ''
  });

  const apiKeyOptions = [
    { value: 'default', label: 'Use Default API Key' },
    { value: 'custom', label: 'Use Custom API Key' }
  ];

  const providerOptions = [
    { value: 'whatsapp', label: 'WhatsApp' },
    { value: 'telegram', label: 'Telegram' },
    { value: 'sms', label: 'SMS' }
  ];

  useEffect(() => {
    fetchDevices();
  }, []);

  const fetchDevices = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/device-settings');
      if (response.ok) {
        const data = await response.json();
        setDevices(data);
      } else {
        toast.error('Failed to fetch devices');
      }
    } catch (error) {
      console.error('Error fetching devices:', error);
      toast.error('Failed to fetch devices');
    } finally {
      setLoading(false);
    }
  };

  const generateDeviceId = () => {
    const timestamp = Date.now().toString();
    const random = Math.random().toString(36).substring(2, 8).toUpperCase();
    const deviceId = `DEV-${timestamp.slice(-6)}-${random}`;
    setCurrentSettings(prev => ({ ...prev, device_id: deviceId }));
  };

  const generateWebhookId = () => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    for (let i = 0; i < 8; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    const webhookUrl = `https://chatbot.growweb.com/chatgpt/SCVTC-S2/${result}`;
    setCurrentSettings(prev => ({ ...prev, webhook_id: webhookUrl }));
  };

  const handleInputChange = (field: keyof DeviceSettings, value: string) => {
    setCurrentSettings(prev => ({ ...prev, [field]: value }));
  };

  const handleSave = async () => {
    try {
      setIsSaving(true);
      const url = editingDevice ? `/api/device-settings/${editingDevice.id}` : '/api/device-settings';
      const method = editingDevice ? 'PUT' : 'POST';
      
      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(currentSettings),
      });

      if (response.ok) {
        toast.success(editingDevice ? 'Device updated successfully!' : 'Device created successfully!');
        setIsDialogOpen(false);
        setEditingDevice(null);
        setCurrentSettings({
          id: '',
          device_id: '',
          api_key_option: 'default',
          webhook_id: '',
          provider: 'whatsapp',
          phone_number: '',
          api_key: '',
          id_device: '',
          id_erp: '',
          id_admin: ''
        });
        fetchDevices();
      } else {
        const errorData = await response.json();
        toast.error(errorData.error || 'Failed to save device');
      }
    } catch (error) {
      console.error('Error saving device:', error);
      toast.error('Failed to save device');
    } finally {
      setIsSaving(false);
    }
  };

  const handleEdit = (device: DeviceSettings) => {
    setEditingDevice(device);
    setCurrentSettings({ ...device });
    setIsDialogOpen(true);
  };

  const handleDelete = async (deviceId: string) => {
    if (!confirm('Are you sure you want to delete this device?')) {
      return;
    }

    try {
      const response = await fetch(`/api/device-settings/${deviceId}`, {
        method: 'DELETE',
      });

      if (response.ok) {
        toast.success('Device deleted successfully!');
        fetchDevices();
      } else {
        toast.error('Failed to delete device');
      }
    } catch (error) {
      console.error('Error deleting device:', error);
      toast.error('Failed to delete device');
    }
  };

  const handleClose = () => {
    setIsDialogOpen(false);
    setEditingDevice(null);
    setCurrentSettings({
      id: '',
      device_id: '',
      api_key_option: 'default',
      webhook_id: '',
      provider: 'whatsapp',
      phone_number: '',
      api_key: '',
      id_device: '',
      id_erp: '',
      id_admin: ''
    });
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success('Copied to clipboard!');
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 space-y-6">
      <Card className="border-slate-200 dark:border-slate-700 shadow-lg">
        <CardHeader className="bg-gradient-to-r from-blue-50 to-purple-50 dark:from-slate-800 dark:to-slate-700">
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-2xl font-bold text-slate-800 dark:text-slate-100 flex items-center gap-2">
                <Settings className="h-6 w-6" />
                Device Settings
              </CardTitle>
            </div>
            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
              <DialogTrigger asChild>
                <Button 
                  className="bg-gradient-to-r from-blue-500 to-purple-600 text-white hover:from-blue-600 hover:to-purple-700 flex items-center gap-2"
                  onClick={() => {
                    setEditingDevice(null);
                    setCurrentSettings({
                      id: '',
                      device_id: '',
                      api_key_option: 'default',
                      webhook_id: '',
                      provider: 'whatsapp',
                      phone_number: '',
                      api_key: '',
                      id_device: '',
                      id_erp: '',
                      id_admin: ''
                    });
                  }}
                >
                  <Plus className="h-4 w-4" />
                  Add New Device
                </Button>
              </DialogTrigger>
              <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                  <DialogTitle className="flex items-center gap-2">
                    <Smartphone className="h-5 w-5" />
                    {editingDevice ? 'Edit Device' : 'Add New Device'}
                  </DialogTitle>
                </DialogHeader>
                <DeviceForm
                  settings={currentSettings}
                  handleInputChange={handleInputChange}
                  generateDeviceId={generateDeviceId}
                  generateWebhookId={generateWebhookId}
                  handleSave={handleSave}
                  handleClose={handleClose}
                  isSaving={isSaving}
                  apiKeyOptions={apiKeyOptions}
                  providerOptions={providerOptions}
                  setSettings={setCurrentSettings}
                />
              </DialogContent>
            </Dialog>
          </div>
        </CardHeader>
        <CardContent className="p-6">
          {devices.length === 0 ? (
            <div className="text-center py-12">
              <Smartphone className="h-12 w-12 text-slate-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-slate-600 dark:text-slate-300 mb-2">No devices configured</h3>
              <p className="text-slate-500 dark:text-slate-400 mb-4">Get started by adding your first device</p>
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
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {devices.map((device) => (
                  <TableRow key={device.id}>
                    <TableCell className="font-mono text-sm">
                      <div className="flex items-center gap-2">
                        <span className="truncate max-w-[150px]">{device.device_id}</span>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => copyToClipboard(device.device_id)}
                          className="h-6 w-6 p-0"
                        >
                          <Copy className="h-3 w-3" />
                        </Button>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="capitalize">
                        {device.provider}
                      </Badge>
                    </TableCell>
                    <TableCell>{device.phone_number}</TableCell>
                    <TableCell>
                      <Badge variant={device.api_key_option === 'custom' ? 'default' : 'secondary'}>
                        {device.api_key_option === 'custom' ? 'Custom' : 'Default'}
                      </Badge>
                    </TableCell>
                    <TableCell>{device.id_device}</TableCell>
                    <TableCell>{device.id_erp}</TableCell>
                    <TableCell>{device.id_admin}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleEdit(device)}
                          className="h-8 w-8 p-0"
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDelete(device.id)}
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