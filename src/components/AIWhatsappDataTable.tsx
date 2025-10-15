import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { useDevice } from '@/contexts/DeviceContext';
import DeviceRequiredPopup from '@/components/DeviceRequiredPopup';
import Swal from 'sweetalert2';
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
const AIWhatsappDataTable = ({ selectedDevice, selectedStage, dateRange }: { 
  selectedDevice?: string; 
  selectedStage?: string;
  dateRange?: { from?: Date; to?: Date };
}) => {
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
  const [deviceFilter, setDeviceFilter] = useState(selectedDevice || 'all');
  const [searchTerm, setSearchTerm] = useState('');
  
  // Update deviceFilter when selectedDevice prop changes
  useEffect(() => {
    setDeviceFilter(selectedDevice || 'all');
  }, [selectedDevice]);
  
  // Available devices for filters (filtered by user's devices)
  const [availableDevices, setAvailableDevices] = useState<string[]>([]);
  
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
        ...(searchTerm && { search: searchTerm })
      });
      
      // Special case: "Welcome Message" means NULL/empty stage in database
      if (selectedStage) {
        if (selectedStage === 'Welcome Message') {
          params.append('stage', 'WELCOME_MESSAGE_NULL');
        } else {
          params.append('stage', selectedStage);
        }
      }
      
      // Add date range filters
      if (dateRange?.from) {
        params.append('startDate', format(dateRange.from, 'yyyy-MM-dd'));
      }
      if (dateRange?.to) {
        params.append('endDate', format(dateRange.to, 'yyyy-MM-dd'));
      }
      
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
        
        // Extract unique devices from the data (already filtered by user's devices)
        const devices = Array.from(new Set(data.data?.map(c => c.id_device).filter(Boolean) || []));
        
        setAvailableDevices(devices);
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
  }, [currentPage, pageSize, deviceFilter, selectedStage, searchTerm, dateRange, has_devices, device_ids]);

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
      Swal.fire({
        icon: 'info',
        title: 'No Data',
        text: 'No data to export',
        confirmButtonColor: '#3b82f6'
      });
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
    const result = await Swal.fire({
      title: 'Are you sure?',
      text: 'Do you want to delete this conversation?',
      icon: 'warning',
      showCancelButton: true,
      confirmButtonColor: '#3085d6',
      cancelButtonColor: '#d33',
      confirmButtonText: 'Yes, delete it!',
      cancelButtonText: 'Cancel'
    });

    if (!result.isConfirmed) {
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
      
      Swal.fire('Deleted!', 'The conversation has been deleted.', 'success');
    } catch (err) {
      console.error('Error deleting conversation:', err);
      Swal.fire('Error!', 'Failed to delete conversation', 'error');
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

      // Refresh data after update - this will trigger summary boxes to refresh too
      fetchAIWhatsappData();
      setShowHumanDialog(false);
      setSelectedProspect(null);
      
      // Force refresh of parent Analytics page if it exists
      window.dispatchEvent(new Event('refreshAnalytics'));
    } catch (err) {
      console.error('Error updating human status:', err);
      alert('Failed to update status');
    }
  };

  const renderConversationHistory = (convLast: any) => {
    if (!convLast) return '-';
    
    try {
      let displayContent;
      
      // Check if it's already parsed JSON
      if (typeof convLast === 'object' && Array.isArray(convLast)) {
        // JSON array format
        displayContent = convLast.map((msg: any, idx: number) => (
          <div key={idx} className="mb-2">
            <span className={`text-xs font-medium ${msg.sender === 'bot' ? 'text-blue-600' : 'text-green-600'}`}>
              {msg.sender === 'bot' ? 'Bot' : 'User'}:
            </span>
            <span className="text-xs ml-1">{msg.message}</span>
          </div>
        ));
      } 
      // Try to parse as JSON
      else if (typeof convLast === 'string') {
        const trimmed = convLast.trim();
        if (trimmed.startsWith('[') || trimmed.startsWith('{')) {
          try {
            const messages = JSON.parse(trimmed);
            if (Array.isArray(messages)) {
              displayContent = messages.map((msg: any, idx: number) => (
                <div key={idx} className="mb-2">
                  <span className={`text-xs font-medium ${msg.sender === 'bot' ? 'text-blue-600' : 'text-green-600'}`}>
                    {msg.sender === 'bot' ? 'Bot' : 'User'}:
                  </span>
                  <span className="text-xs ml-1">{msg.message}</span>
                </div>
              ));
            } else {
              // Plain text display
              displayContent = <div className="text-xs whitespace-pre-wrap">{convLast}</div>;
            }
          } catch (e) {
            // If JSON parse fails, treat as plain text
            displayContent = <div className="text-xs whitespace-pre-wrap">{convLast}</div>;
          }
        } else {
          // Plain text format
          displayContent = <div className="text-xs whitespace-pre-wrap">{convLast}</div>;
        }
      } else {
        return '-';
      }
      
      // Return scrollable div with full content
      return (
        <div className="max-h-20 overflow-y-auto p-1 border rounded text-left">
          {displayContent}
        </div>
      );
    } catch (e) {
      console.error('Error rendering conversation history:', e);
      // Return the raw text if all parsing fails
      return (
        <div className="text-xs max-h-20 overflow-y-auto p-1">
          {typeof convLast === 'string' ? convLast : '-'}
        </div>
      );
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
            <div className="flex-1 min-w-[200px] relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none" />
              <Input
                placeholder="Search by phone number, niche, stage, or marketer..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10"
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
                      <TableHead>Created_at</TableHead>
                      <TableHead>ID Device</TableHead>
                      <TableHead>Phone Number</TableHead>
                      <TableHead>Prospect Name</TableHead>
                      <TableHead>Niche</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Stage</TableHead>
                      <TableHead>Conversation History</TableHead>
                      <TableHead>Keyword Iklan</TableHead>
                      <TableHead>Marketer</TableHead>
                      <TableHead className="text-center">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {conversations.map((conv, index) => (
                      <TableRow key={conv.id_prospect}>
                        <TableCell>{(currentPage - 1) * pageSize + index + 1}</TableCell>
                        <TableCell>
                          {conv.created_at ? format(new Date(conv.created_at), 'dd-MM-yyyy') : '-'}
                        </TableCell>
                        <TableCell>{conv.id_device || '-'}</TableCell>
                        <TableCell>{conv.prospect_num || '-'}</TableCell>
                        <TableCell>{conv.prospect_name || 'Sis'}</TableCell>
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
