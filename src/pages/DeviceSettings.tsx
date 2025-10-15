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
import WablasStatusModal from '@/components/WablasStatusModal';
import WhacenterStatusModal from '@/components/WhacenterStatusModal';
import WahaStatusModal from '@/components/WahaStatusModal';
import Swal from 'sweetalert2';

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
  instance: string;
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
  setSettings: React.Dispatch<React.SetStateAction<DeviceSettings>>;
  isDuplicateIdDevice: boolean;
}> = ({ settings, handleInputChange, generateDeviceId, generateWebhookId, handleSave, handleClose, isSaving, apiKeyOptions, providerOptions, setSettings, isDuplicateIdDevice }) => (
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
          <div className="flex gap-2">
            <Button
              onClick={() => {
                console.log('🔘 GENERATE DEVICE BUTTON CLICKED');
                console.log('📊 Button click timestamp:', new Date().toISOString());
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
        <Input
          value={settings.webhook_id}
          onChange={(e) => handleInputChange('webhook_id', e.target.value)}
          placeholder="https://chatbot.growweb.com/chatgpt/SCVTC-S2/FGcaTDgH"
          className="bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-900 dark:text-slate-100 mt-1"
        />
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
          className={`bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-900 dark:text-slate-100 mt-1 ${
            isDuplicateIdDevice ? 'border-red-500 focus:border-red-500' : ''
          }`}
          required
        />
        {isDuplicateIdDevice && (
          <p className="text-red-500 text-sm mt-1">ID Device already exists! Please enter a unique ID Device.</p>
        )}
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
        disabled={isSaving || !settings.id_device || !settings.id_erp || !settings.id_admin || isDuplicateIdDevice}
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
  const [settings, setSettings] = useState<DeviceSettings>({
    id: '',
    device_id: '',
    api_key_option: 'openai/gpt-4.1',
    webhook_id: '',
    provider: 'wablas',
    phone_number: '',
    api_key: '',
    id_device: '',
    id_erp: '',
    id_admin: '',
    instance: ''
  });
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingDevice, setEditingDevice] = useState<DeviceSettings | null>(null);
  const [statusPopupOpen, setStatusPopupOpen] = useState(false);
  const [selectedDeviceForStatus, setSelectedDeviceForStatus] = useState<DeviceSettings | null>(null);
  const [isDuplicateIdDevice, setIsDuplicateIdDevice] = useState(false);

  const apiKeyOptions = [
    { value: 'openai/gpt-5-chat', label: 'Chat GPT 5o' },
    { value: 'openai/gpt-5-mini', label: 'Chat GPT 5 Mini' },
    { value: 'openai/chatgpt-4o-latest', label: 'Chat GPT 4o' },
    { value: 'openai/gpt-4.1', label: 'Chat GPT 4.1 (NEW)' },
    { value: 'google/gemini-2.5-pro', label: 'GEMINI PRO 2.5' },
    { value: 'google/gemini-pro-1.5', label: 'GEMINI PRO 1.5' }
  ];

  const providerOptions = [
    { value: 'whacenter', label: 'Whacenter' },
    { value: 'wablas', label: 'Wablas' },
    { value: 'waha', label: 'WAHA' }
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
    
    // Check for duplicate ID Device
    if (field === 'id_device') {
      const isDuplicate = devices.some(device => 
        device.id_device === value && 
        device.id !== settings.id // Exclude current device when editing
      );
      setIsDuplicateIdDevice(isDuplicate);
      
      if (isDuplicate) {
        toast.error('ID Device already exists! Please enter a unique ID Device.');
      }
    }
  };

  const handleClose = () => {
    setIsModalOpen(false);
    resetForm();
  };

  const resetForm = () => {
    setSettings({
      id: '',
      device_id: '',
      api_key_option: 'openai/gpt-4.1',
      webhook_id: '',
      provider: 'wablas',
      phone_number: '',
      api_key: '',
      id_device: '',
      id_erp: '',
      id_admin: '',
      instance: ''
    });
    setEditingDevice(null);
    setIsDuplicateIdDevice(false);
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
    const result = await Swal.fire({
      title: 'Are you sure?',
      text: 'Do you want to delete this device?',
      icon: 'warning',
      showCancelButton: true,
      confirmButtonColor: '#3085d6',
      cancelButtonColor: '#d33',
      confirmButtonText: 'Yes, delete it!',
      cancelButtonText: 'Cancel'
    });

    if (!result.isConfirmed) return;
    
    try {
      const response = await fetch(`/api/device-settings/${deviceId}`, {
        method: 'DELETE',
      });
      
      if (response.ok) {
        toast.success('Device deleted successfully!');
        loadDeviceSettings();
        Swal.fire('Deleted!', 'The device has been deleted.', 'success');
      } else {
        throw new Error('Failed to delete device');
      }
    } catch (error) {
      console.error('Error deleting device:', error);
      toast.error('Failed to delete device');
      Swal.fire('Error!', 'Failed to delete device', 'error');
    }
  };

  const handleStatusClick = (device: DeviceSettings) => {
    setSelectedDeviceForStatus(device);
    setStatusPopupOpen(true);
  };

  const handleStatusPopupClose = () => {
    setStatusPopupOpen(false);
    setSelectedDeviceForStatus(null);
  };

  const generateDeviceId = async () => {
    console.log('=== DEVICE GENERATION STARTED ===');
    console.log('🚀 generateDeviceId function called');
    console.log('📋 Current settings state:', JSON.stringify(settings, null, 2));
    
    // Enhanced validation with detailed logging
    console.log('🔍 Starting validation checks...');
    const validationErrors = [];
    
    console.log('📞 Checking phone_number:', {
      value: settings.phone_number,
      type: typeof settings.phone_number,
      trimmed: settings.phone_number?.trim(),
      isEmpty: !settings.phone_number?.trim()
    });
    if (!settings.phone_number?.trim()) {
      console.log('❌ Phone number validation failed');
      validationErrors.push('Phone number is required');
    } else {
      console.log('✅ Phone number validation passed');
    }
    
    console.log('📱 Checking id_device:', {
      value: settings.id_device,
      type: typeof settings.id_device,
      trimmed: settings.id_device?.trim(),
      isEmpty: !settings.id_device?.trim()
    });
    if (!settings.id_device?.trim()) {
      console.log('❌ ID Device validation failed');
      validationErrors.push('ID Device is required');
    } else {
      console.log('✅ ID Device validation passed');
    }
    
    console.log('🏢 Checking provider:', {
      value: settings.provider,
      type: typeof settings.provider,
      trimmed: settings.provider?.trim(),
      isEmpty: !settings.provider?.trim()
    });
    if (!settings.provider?.trim()) {
      console.log('❌ Provider validation failed');
      validationErrors.push('Provider is required');
    } else {
      console.log('✅ Provider validation passed');
    }
    
    console.log('🔑 Checking api_key:', {
      value: settings.api_key ? '[HIDDEN]' : settings.api_key,
      type: typeof settings.api_key,
      length: settings.api_key?.length || 0,
      trimmed_length: settings.api_key?.trim()?.length || 0,
      isEmpty: !settings.api_key?.trim()
    });
    if (!settings.api_key?.trim()) {
      console.log('❌ API Key validation failed');
      validationErrors.push('API Key is required');
    } else {
      console.log('✅ API Key validation passed');
    }
    
    console.log('📊 Validation summary:', {
      totalErrors: validationErrors.length,
      errors: validationErrors
    });
    
    if (validationErrors.length > 0) {
      console.log('🚫 VALIDATION FAILED - Stopping execution');
      console.log('❌ Validation errors:', validationErrors);
      const errorMessage = validationErrors.join(', ');
      console.log('🔔 Showing toast error:', errorMessage);
      
      // Force toast to show
      try {
        toast.error(`Validation Error: ${errorMessage}`);
        console.log('✅ Toast error displayed successfully');
      } catch (toastError) {
        console.error('❌ Failed to show toast:', toastError);
        // Fallback: show browser alert
        alert(`Validation Error: ${errorMessage}`);
      }
      
      return;
    }
    
    console.log('✅ ALL VALIDATIONS PASSED - Proceeding with device generation');

    console.log('Validation passed, starting device generation');
    setIsSaving(true);
    toast.info('Generating device... Please wait');

    try {
      // Get Railway deployment URL from window.location or use environment
      const baseUrl = window.location.origin;
      const webhookUrl = `${baseUrl}/api/webhook/${settings.id_device}`;
      
      let apiResponse;
      
      if (settings.provider === 'whacenter') {
        console.log('🔵 Using Whacenter provider');
        console.log('📡 Preparing Whacenter API request...');
        
        // Whacenter API integration
        const whacenterData = {
          device_name: settings.id_device,
          webhook_url: webhookUrl
        };
        
        const requestBody = {
          ...settings,
          webhook_url: webhookUrl,
          device_data: whacenterData
        };
        
        console.log('📤 Whacenter Request Details:');
        console.log('  - Endpoint: /api/device-settings/generate-whacenter');
        console.log('  - Method: POST');
        console.log('  - Headers: Content-Type: application/json');
        console.log('  - Body:', JSON.stringify(requestBody, null, 2));
        
        const startTime = Date.now();
        
        try {
          console.log('🌐 Making network request to Whacenter API...');
          
          apiResponse = await fetch('/api/device-settings/generate-whacenter', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify(requestBody)
          });
          
          const endTime = Date.now();
          const duration = endTime - startTime;
          
          console.log(`📥 Whacenter API Response (${duration}ms):`);
          console.log('  - Status:', apiResponse.status, apiResponse.statusText);
          console.log('  - OK:', apiResponse.ok);
          console.log('  - Headers:', Object.fromEntries(apiResponse.headers.entries()));
          
          // Clone response to read body without consuming it
          const responseClone = apiResponse.clone();
          const responseText = await responseClone.text();
          console.log('  - Body Length:', responseText.length);
          console.log('  - Body:', responseText);
          
          // Log network timing
          console.log('⏱️ Network Timing:', {
            duration: `${duration}ms`,
            url: '/api/device-settings/generate-whacenter',
            method: 'POST',
            status: apiResponse.status,
            ok: apiResponse.ok
          });
          
        } catch (fetchError) {
          console.error('💥 NETWORK ERROR - Whacenter API Request Failed');
          console.error('❌ Fetch Error Details:', {
            name: fetchError.name,
            message: fetchError.message,
            stack: fetchError.stack,
            cause: fetchError.cause
          });
          console.error('🌐 Network Request Info:', {
            url: '/api/device-settings/generate-whacenter',
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            bodyLength: JSON.stringify(requestBody).length
          });
          throw new Error(`Network Error: ${fetchError.message}`);
        }
      } else if (settings.provider === 'wablas') {
        console.log('🟢 Using Wablas provider');
        console.log('📡 Preparing Wablas API request...');
        
        // Wablas API integration
        const wablasData = {
          device_name: settings.id_device,
          webhook_url: webhookUrl
        };
        
        const requestBody = {
          ...settings,
          webhook_url: webhookUrl,
          device_data: wablasData
        };
        
        console.log('📤 Wablas Request Details:');
        console.log('  - Endpoint: /api/device-settings/generate-wablas');
        console.log('  - Method: POST');
        console.log('  - Headers: Content-Type: application/json');
        console.log('  - Body:', JSON.stringify(requestBody, null, 2));
        
        const startTime = Date.now();
        
        try {
          console.log('🌐 Making network request to Wablas API...');
          
          apiResponse = await fetch('/api/device-settings/generate-wablas', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify(requestBody)
          });
          
          const endTime = Date.now();
          const duration = endTime - startTime;
          
          console.log(`📥 Wablas API Response (${duration}ms):`);
          console.log('  - Status:', apiResponse.status, apiResponse.statusText);
          console.log('  - OK:', apiResponse.ok);
          console.log('  - Headers:', Object.fromEntries(apiResponse.headers.entries()));
          
          // Clone response to read body without consuming it
          const responseClone = apiResponse.clone();
          const responseText = await responseClone.text();
          console.log('  - Body Length:', responseText.length);
          console.log('  - Body:', responseText);
          
          // Log network timing
          console.log('⏱️ Network Timing:', {
            duration: `${duration}ms`,
            url: '/api/device-settings/generate-wablas',
            method: 'POST',
            status: apiResponse.status,
            ok: apiResponse.ok
          });
          
        } catch (fetchError) {
          console.error('💥 NETWORK ERROR - Wablas API Request Failed');
          console.error('❌ Fetch Error Details:', {
            name: fetchError.name,
            message: fetchError.message,
            stack: fetchError.stack,
            cause: fetchError.cause
          });
          console.error('🌐 Network Request Info:', {
            url: '/api/device-settings/generate-wablas',
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            bodyLength: JSON.stringify(requestBody).length
          });
          throw new Error(`Network Error: ${fetchError.message}`);
        }
      } else if (settings.provider === 'waha') {
        console.log('🟢 Using WAHA provider');
        console.log('📡 Preparing WAHA API request...');
        
        // WAHA API integration
        const wahaData = {
          device_name: settings.id_device,
          webhook_url: webhookUrl,
          api_base: 'https://waha-plus-production-705f.up.railway.app/',
          api_key: 'dckr_pat_vxeqEu_CqRi5O3CBHnD7FxhnBz0'
        };
        
        const requestBody = {
          ...settings,
          webhook_url: webhookUrl,
          device_data: wahaData
        };
        
        console.log('📤 WAHA Request Details:');
        console.log('  - Endpoint: /api/device-settings/generate-waha');
        console.log('  - Method: POST');
        console.log('  - Headers: Content-Type: application/json');
        console.log('  - Body:', JSON.stringify(requestBody, null, 2));
        
        const startTime = Date.now();
        
        try {
          console.log('🌐 Making network request to WAHA API...');
          
          apiResponse = await fetch('/api/device-settings/generate-waha', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify(requestBody)
          });
          
          const endTime = Date.now();
          const duration = endTime - startTime;
          
          console.log(`📥 WAHA API Response (${duration}ms):`);
          console.log('  - Status:', apiResponse.status, apiResponse.statusText);
          console.log('  - OK:', apiResponse.ok);
          console.log('  - Headers:', Object.fromEntries(apiResponse.headers.entries()));
          
          // Clone response to read body without consuming it
          const responseClone = apiResponse.clone();
          const responseText = await responseClone.text();
          console.log('  - Body Length:', responseText.length);
          console.log('  - Body:', responseText);
          
          // Log network timing
          console.log('⏱️ Network Timing:', {
            duration: `${duration}ms`,
            url: '/api/device-settings/generate-waha',
            method: 'POST',
            status: apiResponse.status,
            ok: apiResponse.ok
          });
          
        } catch (fetchError) {
          console.error('💥 NETWORK ERROR - WAHA API Request Failed');
          console.error('❌ Fetch Error Details:', {
            name: fetchError.name,
            message: fetchError.message,
            stack: fetchError.stack,
            cause: fetchError.cause
          });
          console.error('🌐 Network Request Info:', {
            url: '/api/device-settings/generate-waha',
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            bodyLength: JSON.stringify(requestBody).length
          });
          throw new Error(`Network Error: ${fetchError.message}`);
        }
      } else {
        // Fallback: Generate local device ID
        const deviceId = `C${Math.floor(Math.random() * 10000).toString().padStart(4, '0')}H`;
        handleInputChange('device_id', deviceId);
        handleInputChange('webhook_id', webhookUrl);
        toast.success('Device ID generated successfully!');
        setIsSaving(false);
        return;
      }

      console.log('🔍 Processing API response...');
      
      if (apiResponse && apiResponse.ok) {
        console.log('✅ API response is OK, parsing JSON...');
        try {
          const result = await apiResponse.json();
          console.log('📋 Parsed API result:', JSON.stringify(result, null, 2));
          
          if (result.success) {
            console.log('🎉 API call successful! Updating form data...');
            
            // Track what data was received
            const receivedData = {
              device_id: result.data?.device_id,
              webhook_url: result.data?.webhook_url,
              api_key: result.data?.api_key,
              provider: result.data?.provider
            };
            console.log('📊 Received data from provider:', receivedData);
            
            // Update form with generated data
            if (result.data?.device_id) {
              console.log('✏️ Setting device_id:', result.data.device_id);
              handleInputChange('device_id', result.data.device_id);
            }
            if (result.data?.webhook_url) {
              console.log('✏️ Setting webhook_id:', result.data.webhook_url);
              handleInputChange('webhook_id', result.data.webhook_url);
            }
            if (result.data?.api_key) {
              console.log('✏️ Setting api_key:', result.data.api_key);
              handleInputChange('api_key', result.data.api_key);
            }
            
            console.log('🎊 Showing success toast');
            toast.success(`Device generated successfully via ${settings.provider}!`);
            
            // Close modal and reset form after successful generation
            console.log('🔄 Closing modal and resetting form...');
            setIsModalOpen(false);
            resetForm();
            
            // Refresh the device settings table to show the new data
            console.log('🔄 Refreshing device settings table...');
            loadDeviceSettings();
            
            console.log('=== DEVICE GENERATION COMPLETED SUCCESSFULLY ===');
          } else {
            console.error('❌ API call failed with message:', result.message);
            console.error('❌ Full error response:', result);
            toast.error(`Provider Error: ${result.message || 'Unknown error from provider'}`);
            throw new Error(result.message || 'Failed to generate device');
          }
        } catch (jsonError) {
          console.error('❌ Failed to parse JSON response:', jsonError);
          const responseText = await apiResponse.text();
          console.error('❌ Raw response text:', responseText);
          toast.error('Invalid response format from server');
          throw new Error('Invalid response format');
        }
      } else {
        console.error('❌ API response not OK');
        console.error('❌ Status:', apiResponse?.status, apiResponse?.statusText);
        
        try {
          const errorText = await apiResponse?.text();
          console.error('❌ Error response body:', errorText);
          
          // Try to parse error as JSON for better error messages
          try {
            const errorJson = JSON.parse(errorText);
            const errorMessage = errorJson.message || errorJson.error || 'Unknown server error';
            toast.error(`Server Error (${apiResponse?.status}): ${errorMessage}`);
          } catch {
            toast.error(`Server Error (${apiResponse?.status}): ${errorText || 'Unknown error'}`);
          }
        } catch (readError) {
          console.error('❌ Failed to read error response:', readError);
          toast.error(`Network Error: Failed to communicate with server (${apiResponse?.status})`);
        }
        
        throw new Error(`HTTP ${apiResponse?.status}: Failed to communicate with device provider`);
      }
    } catch (error) {
      console.error('💥 DEVICE GENERATION FAILED');
      console.error('❌ Error generating device:', error);
      console.error('❌ Error details:', {
        message: error.message,
        stack: error.stack,
        name: error.name,
        provider: settings.provider,
        timestamp: new Date().toISOString()
      });
      
      // Show user-friendly error message
      const userErrorMessage = error.message || 'Unknown error occurred';
      toast.error(`Device Generation Failed: ${userErrorMessage}`);
      
      // Fallback: Generate local device ID
      console.log('🔄 Generating fallback local device ID...');
      const deviceId = `C${Math.floor(Math.random() * 10000).toString().padStart(4, '0')}H`;
      console.log('🆔 Generated fallback device ID:', deviceId);
      handleInputChange('device_id', deviceId);
      toast.info('Generated local device ID as fallback');
      console.log('=== FALLBACK DEVICE GENERATION COMPLETED ===');
    } finally {
      console.log('🏁 Device generation process completed, setting isSaving to false');
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
              setSettings={setSettings}
              isDuplicateIdDevice={isDuplicateIdDevice}
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
                  <TableHead>NO</TableHead>
                  <TableHead>ID</TableHead>
                  <TableHead>ID DEVICE</TableHead>
                  <TableHead>PHONE NUMBER</TableHead>
                  <TableHead>ID ERP</TableHead>
                  <TableHead>ID ADMIN</TableHead>
                  <TableHead>PROVIDER</TableHead>
                  <TableHead>INSTANCE</TableHead>
                  <TableHead>WEBHOOK ID</TableHead>
                  <TableHead>API KEY OPTION</TableHead>
                  <TableHead>API KEY</TableHead>
                  <TableHead className="text-center">STATUS DEVICE</TableHead>
                  <TableHead className="text-right">ACTIONS</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {devices.map((device, index) => (
                  <TableRow key={device.id}>
                    <TableCell className="font-medium">{index + 1}</TableCell>
                    <TableCell className="font-medium">{device.id}</TableCell>
                    <TableCell>{device.id_device}</TableCell>
                    <TableCell>{device.phone_number || 'Not set'}</TableCell>
                    <TableCell>{device.id_erp}</TableCell>
                    <TableCell>{device.id_admin}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{device.provider}</Badge>
                    </TableCell>
                    <TableCell>{device.instance || 'Not set'}</TableCell>
                    <TableCell>{device.webhook_id || 'Not set'}</TableCell>
                    <TableCell>{apiKeyOptions.find(opt => opt.value === device.api_key_option)?.label}</TableCell>
                    <TableCell>{device.api_key ? '***' : 'Not set'}</TableCell>
                    <TableCell>
                      <Badge 
                        variant={device.device_id ? "default" : "secondary"}
                        className="cursor-pointer hover:opacity-80 transition-opacity"
                        onClick={() => handleStatusClick(device)}
                      >
                        Status Device
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
      
      {/* Provider-specific status modals */}
      {selectedDeviceForStatus?.provider === 'wablas' && (
        <WablasStatusModal
          isOpen={statusPopupOpen}
          onClose={handleStatusPopupClose}
          deviceId={selectedDeviceForStatus?.id || ''}
          deviceName={selectedDeviceForStatus?.device_id}
        />
      )}
      
      {selectedDeviceForStatus?.provider === 'whacenter' && (
        <WhacenterStatusModal
          isOpen={statusPopupOpen}
          onClose={handleStatusPopupClose}
          deviceId={selectedDeviceForStatus?.id || ''}
          deviceName={selectedDeviceForStatus?.device_id}
        />
      )}
      
      {selectedDeviceForStatus?.provider === 'waha' && (
        <WahaStatusModal
          isOpen={statusPopupOpen}
          onClose={handleStatusPopupClose}
          deviceId={selectedDeviceForStatus?.id || ''}
          deviceName={selectedDeviceForStatus?.device_id}
        />
      )}
    </div>
  );
};



export default DeviceSettings;