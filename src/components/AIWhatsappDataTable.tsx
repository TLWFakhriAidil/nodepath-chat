import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { useDevice } from '@/contexts/DeviceContext';
import DeviceRequiredPopup from '@/components/DeviceRequiredPopup';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { 
  Search, 
  RefreshCw, 
  ChevronLeft, 
  ChevronRight,
  MessageSquare,
  User,
  Calendar,
  Filter,
  Download,
  Trash2,
  Bot,
  UserCheck,
  ExternalLink
} from 'lucide-react';
import { format } from 'date-fns';

// AI WhatsApp conversation interface
interface AIWhatsappConversation {
  id_prospect: string;
  id_device: string;
  prospect_num: string;
  prospect_name: string;
  stage: string;
  date_order: string;
  conv_last: any;
  conv_current: string;
  human: number;
  niche: string;
  jam: string;
  intro: string;
  catatan_staff: string;
  balas: string;
  data_image: string;
  conv_stage: string;
  bot_balas: string;
  keywordiklan: string;
  marketer: string;
  update_today: string;
  created_at: string;
  updated_at: string;
}

// API response interface
interface AIWhatsappDataResponse {
  success: boolean;
  data: AIWhatsappConversation[];
  pagination: {
    current_page: number;
    limit: number;
    total_records: number;
    total_pages: number;
  };
}

/**
 * AI WhatsApp Data Table component with device-based filtering
 * Automatically filters conversations by user's configured devices
 */
const AIWhatsappDataTable = () => {
  const { has_devices, device_ids } = useDevice();
  const [conversations, setConversations] = useState<AIWhatsappConversation[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showDeviceRequiredPopup, setShowDeviceRequiredPopup] = useState(false);
  
  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalRecords, setTotalRecords] = useState(0);
  const [pageSize, setPageSize] = useState(10);
  
  // Filter state
  const [deviceFilter, setDeviceFilter] = useState('all');
  const [stageFilter, setStageFilter] = useState('all');
  const [searchTerm, setSearchTerm] = useState('');
  
  // Available devices and stages for filters (filtered by user's devices)
  const [availableDevices, setAvailableDevices] = useState<string[]>([]);
  const [availableStages, setAvailableStages] = useState<string[]>([]);
  
  // Dialog state for human/AI toggle
  const [showHumanDialog, setShowHumanDialog] = useState(false);
  const [selectedProspect, setSelectedProspect] = useState<{id: string, human: number, name: string} | null>(null);
  const [selectedHumanStatus, setSelectedHumanStatus] = useState<'AI' | 'Human'>('AI');

  /**
   * Fetch AI WhatsApp data from the backend
   * Automatically filters by user's device IDs from device context
   */
  const fetchAIWhatsappData = async () => {
    console.log('AIWhatsappDataTable: Fetching data...');
    console.log('AIWhatsappDataTable: Device IDs from context:', device_ids);
    
    setLoading(true);
    setError(null);
    
    try {
      const params = new URLSearchParams({
        page: currentPage.toString(),
        limit: pageSize.toString(),
        ...(deviceFilter && deviceFilter !== 'all' && { device_id: deviceFilter }),
        ...(stageFilter && stageFilter !== 'all' && { stage: stageFilter }),
        ...(searchTerm && { search: searchTerm })
      });
      
      // Add user's device IDs to filter the data
      if (device_ids && device_ids.length > 0) {
        params.append('user_device_ids', device_ids.join(','));
      }
      
      const response = await fetch(`/api/ai-whatsapp/ai/ai-whatsapp/data?${params}`, {
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      
      const data: AIWhatsappDataResponse = await response.json();
      console.log('AIWhatsappDataTable: Received data:', data);
      
      if (data.success) {
        setConversations(data.data || []);
        setTotalPages(data.pagination?.total_pages || 1);
        setTotalRecords(data.pagination?.total_records || 0);
        
        // Extract unique devices and stages from the data (already filtered by user's devices)
        const devices = Array.from(new Set(data.data?.map(c => c.id_device).filter(Boolean) || []));
        const stages = Array.from(new Set(data.data?.map(c => c.stage || 'Welcome Message').filter(Boolean) || []));
        
        setAvailableDevices(devices);
        setAvailableStages(stages);
      } else {
        throw new Error('Failed to fetch data');
      }
    } catch (err) {
      console.error('AIWhatsappDataTable: Error fetching data:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch AI WhatsApp data');
      setConversations([]);
    } finally {
      setLoading(false);
    }
  };

  // Fetch data on component mount and when filters/pagination change
  useEffect(() => {
    if (has_devices && device_ids && device_ids.length > 0) {
      fetchAIWhatsappData();
    } else if (!has_devices) {
      setShowDeviceRequiredPopup(true);
    }
  }, [currentPage, pageSize, deviceFilter, stageFilter, searchTerm, has_devices, device_ids]);

  const handlePageChange = (newPage: number) => {
    if (newPage >= 1 && newPage <= totalPages) {
      setCurrentPage(newPage);
    }
  };

  const handleRefresh = () => {
    fetchAIWhatsappData();
  };

  const handleExport = () => {
    // Export filtered data to CSV
    if (conversations.length === 0) {
      alert('No data to export');
      return;
    }

    const headers = ['No', 'ID Device', 'Phone Number', 'Prospect Name', 'Niche', 'Status', 'Stage', 'Keyword Iklan', 'Marketer', 'Updated'];
    const csvData = conversations.map((conv, index) => [
      index + 1,
      conv.id_device || '',
      conv.prospect_num || '',
      conv.prospect_name || '',
      conv.niche || '',
      conv.human === 1 ? 'Human' : 'AI',
      conv.stage || 'Welcome Message',
      conv.keywordiklan || '',
      conv.marketer || '',
      conv.updated_at ? format(new Date(conv.updated_at), 'dd/MM/yyyy HH:mm') : ''
    ]);

    const csvContent = [
      headers.join(','),
      ...csvData.map(row => row.map(cell => {
        const cellStr = String(cell);
        if (cellStr.includes(',') || cellStr.includes('"') || cellStr.includes('\n')) {
          return `"${cellStr.replace(/"/g, '""')}"`;
        }
        return cellStr;
      }).join(','))
    ].join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    const filename = `ai_whatsapp_export_${format(new Date(), 'yyyy-MM-dd')}.csv`;
    
    link.setAttribute('href', url);
    link.setAttribute('download', filename);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this conversation?')) {
      return;
    }

    try {
      const response = await fetch(`/api/ai-whatsapp/ai/ai-whatsapp/${id}`, {
        method: 'DELETE',
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to delete conversation');
      }

      // Refresh data after deletion
      fetchAIWhatsappData();
    } catch (err) {
      console.error('Error deleting conversation:', err);
      alert('Failed to delete conversation');
    }
  };

  const handleHumanToggleClick = (id: string, currentHuman: number, name: string) => {
    setSelectedProspect({ id, human: currentHuman, name });
    setSelectedHumanStatus(currentHuman === 1 ? 'Human' : 'AI');
    setShowHumanDialog(true);
  };

  const handleHumanToggleConfirm = async () => {
    if (!selectedProspect) return;

    const newHumanValue = selectedHumanStatus === 'Human' ? 1 : 0;

    try {
      const response = await fetch(`/api/ai-whatsapp/ai/ai-whatsapp/${selectedProspect.id}/human`, {
        method: 'PUT',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ human: newHumanValue }),
      });

      if (!response.ok) {
        throw new Error('Failed to update status');
      }

      // Refresh data after update
      fetchAIWhatsappData();
      setShowHumanDialog(false);
      setSelectedProspect(null);
    } catch (err) {
      console.error('Error updating human status:', err);
      alert('Failed to update status');
    }
  };

  const renderConversationHistory = (convLast: any) => {
    if (!convLast) return '-';
    
    try {
      // If it's already an object, use it directly
      const messages = typeof convLast === 'string' ? JSON.parse(convLast) : convLast;
      
      if (!Array.isArray(messages) || messages.length === 0) return '-';
      
      // Get last 2 messages
      const lastMessages = messages.slice(-2);
      
      return (
        <Dialog>
          <DialogTrigger asChild>
            <Button variant="ghost" size="sm" className="h-auto p-1">
              <div className="text-left max-w-xs">
                {lastMessages.map((msg: any, idx: number) => (
                  <div key={idx} className="mb-1">
                    <span className={`text-xs font-medium ${msg.sender === 'bot' ? 'text-blue-600' : 'text-green-600'}`}>
                      {msg.sender === 'bot' ? 'Bot' : 'User'}:
                    </span>
                    <span className="text-xs ml-1 line-clamp-1">{msg.message}</span>
                  </div>
                ))}
              </div>
              <ExternalLink className="w-3 h-3 ml-1" />
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Conversation History</DialogTitle>
            </DialogHeader>
            <div className="space-y-2">
              {messages.map((msg: any, idx: number) => (
                <div key={idx} className={`p-3 rounded-lg ${msg.sender === 'bot' ? 'bg-blue-50' : 'bg-green-50'}`}>
                  <div className="flex justify-between items-start mb-1">
                    <span className={`font-medium ${msg.sender === 'bot' ? 'text-blue-700' : 'text-green-700'}`}>
                      {msg.sender === 'bot' ? 'Bot' : 'User'}
                    </span>
                    {msg.timestamp && (
                      <span className="text-xs text-gray-500">
                        {format(new Date(msg.timestamp), 'dd/MM HH:mm')}
                      </span>
                    )}
                  </div>
                  <p className="text-sm whitespace-pre-wrap">{msg.message}</p>
                </div>
              ))}
            </div>
          </DialogContent>
        </Dialog>
      );
    } catch (e) {
      console.error('Error parsing conversation history:', e);
      return '-';
    }
  };

  if (!has_devices) {
    return (
      <div>
        <DeviceRequiredPopup 
          open={showDeviceRequiredPopup} 
          onOpenChange={setShowDeviceRequiredPopup} 
        />
        <Card>
          <CardContent className="text-center py-12">
            <MessageSquare className="w-16 h-16 mx-auto text-muted-foreground mb-4" />
            <h2 className="text-2xl font-semibold mb-2">No Devices Configured</h2>
            <p className="text-muted-foreground">
              Please configure at least one device to view AI WhatsApp conversations
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <Card className="border-0 shadow-lg">
        <CardHeader>
          <div className="flex justify-between items-center">
            <CardTitle className="flex items-center space-x-2">
              <MessageSquare className="w-5 h-5" />
              <span>AI WhatsApp Conversations</span>
            </CardTitle>
            <div className="flex items-center space-x-2">
              <Button onClick={handleExport} variant="outline" size="sm">
                <Download className="w-4 h-4 mr-2" />
                Export
              </Button>
              <Button onClick={handleRefresh} variant="outline" size="sm">
                <RefreshCw className="w-4 h-4 mr-2" />
                Refresh
              </Button>
            </div>
          </div>
          <div className="text-sm text-muted-foreground mt-2">
            View and manage all AI WhatsApp conversation records
          </div>
        </CardHeader>
        <CardContent>
          {/* Filters */}
          <div className="flex flex-wrap gap-4 mb-6">
            <div className="flex-1 min-w-[200px]">
              <Input
                placeholder="Search by phone number, niche, stage, or marketer..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full"
                icon={<Search className="w-4 h-4" />}
              />
            </div>
            {availableDevices.length > 1 && (
              <Select value={deviceFilter} onValueChange={setDeviceFilter}>
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="All Devices" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Devices</SelectItem>
                  {availableDevices.map(device => (
                    <SelectItem key={device} value={device}>{device}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
            <Select value={stageFilter} onValueChange={setStageFilter}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="All Stages" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Stages</SelectItem>
                {availableStages.map(stage => (
                  <SelectItem key={stage} value={stage}>{stage}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Data Table */}
          {loading ? (
            <div className="text-center py-12">
              <RefreshCw className="w-8 h-8 animate-spin mx-auto mb-4" />
              <p className="text-muted-foreground">Loading conversations...</p>
            </div>
          ) : error ? (
            <div className="text-center py-12">
              <p className="text-red-600 mb-4">{error}</p>
              <Button onClick={handleRefresh}>Try Again</Button>
            </div>
          ) : conversations.length === 0 ? (
            <div className="text-center py-12">
              <MessageSquare className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
              <p className="text-muted-foreground">No conversations found</p>
            </div>
          ) : (
            <>
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-12">No</TableHead>
                      <TableHead>ID Device</TableHead>
                      <TableHead>Phone Number</TableHead>
                      <TableHead>Prospect Name</TableHead>
                      <TableHead>Niche</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Stage</TableHead>
                      <TableHead>Conversation History</TableHead>
                      <TableHead>Keyword Iklan</TableHead>
                      <TableHead>Marketer</TableHead>
                      <TableHead>Updated</TableHead>
                      <TableHead className="text-center">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {conversations.map((conv, index) => (
                      <TableRow key={conv.id_prospect}>
                        <TableCell>{(currentPage - 1) * pageSize + index + 1}</TableCell>
                        <TableCell>{conv.id_device || '-'}</TableCell>
                        <TableCell>{conv.prospect_num || '-'}</TableCell>
                        <TableCell>{conv.prospect_name || '-'}</TableCell>
                        <TableCell>{conv.niche || '-'}</TableCell>
                        <TableCell>
                          <Badge 
                            variant={conv.human === 1 ? "secondary" : "default"}
                            className="cursor-pointer"
                            onClick={() => handleHumanToggleClick(conv.id_prospect, conv.human, conv.prospect_name)}
                          >
                            {conv.human === 1 ? (
                              <>
                                <UserCheck className="w-3 h-3 mr-1" />
                                Human
                              </>
                            ) : (
                              <>
                                <Bot className="w-3 h-3 mr-1" />
                                AI
                              </>
                            )}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">
                            {conv.stage || 'Welcome Message'}
                          </Badge>
                        </TableCell>
                        <TableCell>{renderConversationHistory(conv.conv_last)}</TableCell>
                        <TableCell>{conv.keywordiklan || '-'}</TableCell>
                        <TableCell>{conv.marketer || '-'}</TableCell>
                        <TableCell>
                          {conv.updated_at ? format(new Date(conv.updated_at), 'dd/MM HH:mm') : '-'}
                        </TableCell>
                        <TableCell>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleDelete(conv.id_prospect)}
                            className="text-red-600 hover:text-red-700"
                          >
                            <Trash2 className="w-4 h-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>

              {/* Pagination */}
              <div className="flex items-center justify-between mt-4">
                <div className="text-sm text-muted-foreground">
                  Showing {(currentPage - 1) * pageSize + 1} to {Math.min(currentPage * pageSize, totalRecords)} of {totalRecords} records
                </div>
                <div className="flex items-center space-x-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handlePageChange(currentPage - 1)}
                    disabled={currentPage === 1}
                  >
                    <ChevronLeft className="w-4 h-4" />
                    Previous
                  </Button>
                  <div className="flex items-center space-x-1">
                    <span className="text-sm">Page</span>
                    <Input
                      type="number"
                      min="1"
                      max={totalPages}
                      value={currentPage}
                      onChange={(e) => {
                        const page = parseInt(e.target.value);
                        if (!isNaN(page)) {
                          handlePageChange(page);
                        }
                      }}
                      className="w-16 text-center"
                    />
                    <span className="text-sm">of {totalPages}</span>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handlePageChange(currentPage + 1)}
                    disabled={currentPage === totalPages}
                  >
                    Next
                    <ChevronRight className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Human/AI Toggle Dialog */}
      <Dialog open={showHumanDialog} onOpenChange={setShowHumanDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Change Status</DialogTitle>
            <DialogDescription>
              Change the conversation status for {selectedProspect?.name || 'this prospect'}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Select Status:</label>
              <Select value={selectedHumanStatus} onValueChange={(value: 'AI' | 'Human') => setSelectedHumanStatus(value)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="AI">
                    <div className="flex items-center">
                      <Bot className="w-4 h-4 mr-2" />
                      AI (Automated responses)
                    </div>
                  </SelectItem>
                  <SelectItem value="Human">
                    <div className="flex items-center">
                      <UserCheck className="w-4 h-4 mr-2" />
                      Human (Manual takeover)
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowHumanDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleHumanToggleConfirm}>
              Confirm
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default AIWhatsappDataTable;
