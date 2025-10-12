import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { DatePickerWithRange } from '@/components/ui/date-picker';
import { useDevice } from '@/contexts/DeviceContext';
// DeviceRequiredPopup removed - DeviceRequiredWrapper handles this at a higher level
import Swal from 'sweetalert2';
import { 
  BarChart3, 
  TrendingUp, 
  Users, 
  MessageSquare, 
  Clock, 
  Target,
  Download,
  Calendar,
  Filter,
  RefreshCw,
  ArrowUpRight,
  ArrowDownRight,
  Activity,
  Zap,
  Globe,
  Search,
  ChevronLeft,
  ChevronRight,
  Trash2,
  Bot,
  UserCheck
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { DateRange } from 'react-day-picker';
import { format } from 'date-fns';

// Conversation interface
interface Conversation {
  id_prospect: string;
  id_device: string;
  prospect_num: string;
  prospect_name: string;
  stage: string;
  human: number;
  niche: string;
  keywordiklan: string;
  marketer: string;
  created_at: string;
  conv_last: any;
}

const Analytics = () => {
  const { device_ids } = useDevice(); // Remove has_devices since DeviceRequiredWrapper handles this
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // All conversations data (single source of truth)
  const [allConversations, setAllConversations] = useState<Conversation[]>([]);
  
  // Filters
  const [dateRange, setDateRange] = useState<DateRange | undefined>({
    from: new Date(new Date().getFullYear(), new Date().getMonth(), 1),
    to: new Date()
  });
  const [selectedDevice, setSelectedDevice] = useState<string>('');
  const [selectedStage, setSelectedStage] = useState<string>('');
  const [searchTerm, setSearchTerm] = useState('');
  
  // Pagination
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(10);
  
  // Dialog
  const [showHumanDialog, setShowHumanDialog] = useState(false);
  const [selectedProspect, setSelectedProspect] = useState<{id: string, human: number, name: string} | null>(null);
  const [selectedHumanStatus, setSelectedHumanStatus] = useState<'AI' | 'Human'>('AI');
  
  // Fetch all data once
  const fetchAllData = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const params = new URLSearchParams({
        page: '1',
        limit: '10000' // Get all records
      });
      
      if (dateRange?.from) {
        params.append('startDate', format(dateRange.from, 'yyyy-MM-dd'));
      }
      if (dateRange?.to) {
        params.append('endDate', format(dateRange.to, 'yyyy-MM-dd'));
      }
      
      if (selectedDevice) {
        params.append('device_id', selectedDevice);
      } else if (device_ids && device_ids.length > 0) {
        params.append('user_device_ids', device_ids.join(','));
      }
      
      // NOTE: Stage filtering is done client-side (see filteredConversations below)
      // Do NOT send stage parameter to backend as it doesn't handle it correctly
      
      if (searchTerm) {
        params.append('search', searchTerm);
      }
      
      const response = await fetch(`/api/ai-whatsapp/ai/ai-whatsapp/data?${params}`, {
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      });
      
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
      
      const data = await response.json();
      setAllConversations(data.data || []);
    } catch (err) {
      console.error('Error fetching data:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch data');
      setAllConversations([]);
    } finally {
      setLoading(false);
    }
  };
  
  useEffect(() => {
    fetchAllData(); // No need to check has_devices since DeviceRequiredWrapper handles this
  }, [dateRange, selectedDevice, searchTerm, device_ids]); // Note: selectedStage removed - it's client-side only
  
  // Client-side filtering (since backend doesn't filter properly)
  const filteredConversations = allConversations.filter(conv => {
    // Date filtering
    if (dateRange?.from && dateRange?.to) {
      const convDate = new Date(conv.created_at);
      const startDate = new Date(dateRange.from);
      const endDate = new Date(dateRange.to);
      
      // Set time to start/end of day for accurate comparison
      startDate.setHours(0, 0, 0, 0);
      endDate.setHours(23, 59, 59, 999);
      
      if (convDate < startDate || convDate > endDate) {
        return false;
      }
    }
    
    // Stage filtering (including Welcome Message for NULL/empty stages)
    if (selectedStage) {
      if (selectedStage === 'Welcome Message') {
        // Filter for records where stage is null, empty, or 'Welcome Message'
        if (conv.stage && conv.stage !== '' && conv.stage !== 'Welcome Message') {
          return false;
        }
      } else {
        // Filter for specific stage
        if (conv.stage !== selectedStage) {
          return false;
        }
      }
    }
    
    return true;
  });
  
  // Calculate statistics from filteredConversations (not allConversations)
  const stats = {
    totalConversations: filteredConversations.length,
    aiActiveConversations: filteredConversations.filter(c => c.human === 0).length,
    humanTakeovers: filteredConversations.filter(c => c.human === 1).length,
    uniqueDevices: new Set(filteredConversations.map(c => c.id_device)).size,
    uniqueNiches: new Set(filteredConversations.map(c => c.niche).filter(Boolean)).size,
    conversationsWithStages: filteredConversations.filter(c => c.stage && c.stage !== '').length,
  };
  
  // Get available stages from ALL conversations (not filtered) for dropdown
  const allStages = allConversations.reduce((acc, conv) => {
    const stage = conv.stage || 'Welcome Message';
    acc[stage] = (acc[stage] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);
  const availableStages = Object.keys(allStages);
  
  // Calculate stage distribution from filtered data (for chart display)
  const stageDistribution = filteredConversations.reduce((acc, conv) => {
    const stage = conv.stage || 'Welcome Message';
    acc[stage] = (acc[stage] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);
  
  // Calculate daily breakdown from filtered data
  const dailyBreakdown = filteredConversations.reduce((acc, conv) => {
    const date = format(new Date(conv.created_at), 'yyyy-MM-dd');
    acc[date] = (acc[date] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);
  
  const dailyBreakdownArray = Object.entries(dailyBreakdown)
    .map(([date, conversations]) => ({ date, conversations }))
    .sort((a, b) => a.date.localeCompare(b.date));
  
  // Paginate filtered conversations for table
  const startIndex = (currentPage - 1) * pageSize;
  const endIndex = startIndex + pageSize;
  const paginatedConversations = filteredConversations.slice(startIndex, endIndex);
  const totalPages = Math.ceil(filteredConversations.length / pageSize);
  
  const handleRefresh = () => {
    fetchAllData();
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

    if (!result.isConfirmed) return;

    try {
      const response = await fetch(`/api/ai-whatsapp/ai/ai-whatsapp/${id}`, {
        method: 'DELETE',
        credentials: 'include',
      });

      if (!response.ok) throw new Error('Failed to delete');
      
      Swal.fire('Deleted!', 'The conversation has been deleted.', 'success');
      fetchAllData();
    } catch (err) {
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
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ human: newHumanValue }),
      });

      if (!response.ok) throw new Error('Failed to update');
      
      setShowHumanDialog(false);
      setSelectedProspect(null);
      fetchAllData();
    } catch (err) {
      alert('Failed to update status');
    }
  };

  const handleExport = () => {
    if (filteredConversations.length === 0) {
      alert('No data to export');
      return;
    }

    // Prepare CSV headers
    const headers = ['No', 'Created At', 'ID Device', 'Phone Number', 'Prospect Name', 'Niche', 'Status', 'Stage', 'Keyword Iklan', 'Marketer'];
    
    // Prepare CSV rows from filtered data
    const csvData = filteredConversations.map((conv, index) => {
      return [
        index + 1,
        conv.created_at ? format(new Date(conv.created_at), 'dd-MM-yyyy HH:mm:ss') : '',
        conv.id_device || '',
        conv.prospect_num || '',
        conv.prospect_name || '',
        conv.niche || '',
        conv.human === 1 ? 'Human' : 'AI',
        conv.stage || 'Welcome Message',
        conv.keywordiklan || '',
        conv.marketer || ''
      ];
    });

    // Convert to CSV string
    const csvContent = [
      headers.join(','),
      ...csvData.map(row => row.map(cell => {
        // Escape quotes and wrap in quotes if contains comma
        const cellStr = String(cell);
        if (cellStr.includes(',') || cellStr.includes('"') || cellStr.includes('\n')) {
          return `"${cellStr.replace(/"/g, '""')}"`;
        }
        return cellStr;
      }).join(','))
    ].join('\n');

    // Create blob and download
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    
    // Generate filename with current date
    const today = new Date();
    const dateStr = today.toISOString().split('T')[0];
    const filename = `chatbot_ai_export_${dateStr}.csv`;
    
    link.setAttribute('href', url);
    link.setAttribute('download', filename);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    
    console.log(`Exported ${filteredConversations.length} records to ${filename}`);
  };
  
  const renderConversationHistory = (convLast: any) => {
    if (!convLast) return '-';
    
    try {
      let displayContent;
      
      if (typeof convLast === 'object' && Array.isArray(convLast)) {
        displayContent = convLast.map((msg: any, idx: number) => (
          <div key={idx} className="mb-2">
            <span className={`text-xs font-medium ${msg.sender === 'bot' ? 'text-blue-600' : 'text-green-600'}`}>
              {msg.sender === 'bot' ? 'Bot' : 'User'}:
            </span>
            <span className="text-xs ml-1">{msg.message}</span>
          </div>
        ));
      } else if (typeof convLast === 'string') {
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
              displayContent = <div className="text-xs whitespace-pre-wrap">{convLast}</div>;
            }
          } catch (e) {
            displayContent = <div className="text-xs whitespace-pre-wrap">{convLast}</div>;
          }
        } else {
          displayContent = <div className="text-xs whitespace-pre-wrap">{convLast}</div>;
        }
      } else {
        return '-';
      }
      
      return (
        <div className="max-h-20 overflow-y-auto p-1 border rounded text-left">
          {displayContent}
        </div>
      );
    } catch (e) {
      return <div className="text-xs max-h-20 overflow-y-auto p-1">{typeof convLast === 'string' ? convLast : '-'}</div>;
    }
  };

  // Device check removed - DeviceRequiredWrapper handles this at App.tsx level
  return (
    <div className="space-y-6">
      {/* Header with all filters */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white mb-2">
            Analytics Dashboard
          </h1>
          <p className="text-slate-600 dark:text-slate-400">
            Monitor your chatbot performance and user engagement
          </p>
        </div>
        
        <div className="flex items-center space-x-2">
          <Input
            placeholder="Search conversations..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-64"
          />
          
          <DatePickerWithRange
            from={dateRange?.from}
            to={dateRange?.to}
            onSelect={setDateRange}
            className="w-auto"
          />
          
          <Button variant="outline" size="sm" onClick={handleRefresh} disabled={loading}>
            <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm">
                <Filter className="w-4 h-4 mr-2" />
                {selectedDevice ? `Device: ${selectedDevice}` : 'All Devices'}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuItem onClick={() => setSelectedDevice('')}>
                All Devices
              </DropdownMenuItem>
              {device_ids && device_ids.map((deviceId) => (
                <DropdownMenuItem key={deviceId} onClick={() => setSelectedDevice(deviceId)}>
                  {deviceId}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
          
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm">
                <Filter className="w-4 h-4 mr-2" />
                {selectedStage ? `Stage: ${selectedStage}` : 'All Stages'}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuItem onClick={() => setSelectedStage('')}>
                All Stages
              </DropdownMenuItem>
              {availableStages.map((stage) => (
                <DropdownMenuItem key={stage} onClick={() => setSelectedStage(stage)}>
                  {stage}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
          
          <Button className="bg-blue-600 hover:bg-blue-700" size="sm" onClick={handleExport}>
            <Download className="w-4 h-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      {/* Loading */}
      {loading && (
        <div className="flex items-center justify-center py-8">
          <RefreshCw className="w-6 h-6 animate-spin mr-2" />
          <span>Loading data...</span>
        </div>
      )}
      
      {/* Error */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-red-800">Error: {error}</p>
          <Button onClick={handleRefresh} className="mt-2" size="sm">Try Again</Button>
        </div>
      )}

      {/* Key Metrics */}
      {!loading && !error && (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <Card className="border-0 shadow-lg">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-slate-600 dark:text-slate-400 mb-1">
                      Total Conversations
                    </p>
                    <p className="text-2xl font-bold text-white">{stats.totalConversations}</p>
                    <p className="text-xs text-slate-400 mt-1">0% vs last period</p>
                  </div>
                  <div className="w-12 h-12 bg-blue-100 dark:bg-blue-900/20 rounded-lg flex items-center justify-center">
                    <MessageSquare className="w-6 h-6 text-blue-600" />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-0 shadow-lg">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-slate-600 dark:text-slate-400 mb-1">
                      AI Active Conversations
                    </p>
                    <p className="text-2xl font-bold text-white">{stats.aiActiveConversations}</p>
                    <p className="text-xs text-slate-400 mt-1">0% vs last period</p>
                  </div>
                  <div className="w-12 h-12 bg-green-100 dark:bg-green-900/20 rounded-lg flex items-center justify-center">
                    <Users className="w-6 h-6 text-green-600" />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-0 shadow-lg">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-slate-600 dark:text-slate-400 mb-1">
                      Human Takeovers
                    </p>
                    <p className="text-2xl font-bold text-white">{stats.humanTakeovers}</p>
                    <p className="text-xs text-slate-400 mt-1">0% vs last period</p>
                  </div>
                  <div className="w-12 h-12 bg-orange-100 dark:bg-orange-900/20 rounded-lg flex items-center justify-center">
                    <Clock className="w-6 h-6 text-orange-600" />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-0 shadow-lg">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-slate-600 dark:text-slate-400 mb-1">
                      Unique Devices
                    </p>
                    <p className="text-2xl font-bold text-white">{stats.uniqueDevices}</p>
                    <p className="text-xs text-slate-400 mt-1">0% vs last period</p>
                  </div>
                  <div className="w-12 h-12 bg-purple-100 dark:bg-purple-900/20 rounded-lg flex items-center justify-center">
                    <Target className="w-6 h-6 text-purple-600" />
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Charts Row */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Daily Conversation Trends */}
            <Card className="border-0 shadow-xl">
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <BarChart3 className="w-5 h-5 text-blue-600" />
                    <span>Daily Conversation Trends</span>
                  </div>
                  <Badge variant="secondary">Daily</Badge>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {dailyBreakdownArray.length > 0 ? (
                    dailyBreakdownArray.map((day) => {
                      const maxConversations = Math.max(...dailyBreakdownArray.map(d => d.conversations));
                      const conversationWidth = maxConversations > 0 ? (day.conversations / maxConversations) * 100 : 0;
                      
                      return (
                        <div key={day.date} className="space-y-2">
                          <div className="flex items-center justify-between text-sm">
                            <span className="text-slate-600 dark:text-slate-400">
                              {new Date(day.date).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })}
                            </span>
                            <span className="text-slate-900 dark:text-white font-medium">
                              {day.conversations} conversations
                            </span>
                          </div>
                          <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2">
                            <div 
                              className="bg-blue-600 h-2 rounded-full transition-all duration-300" 
                              style={{ width: `${conversationWidth}%` }}
                            />
                          </div>
                        </div>
                      );
                    })
                  ) : (
                    <div className="text-center py-8 text-slate-500">No daily data available</div>
                  )}
                </div>
              </CardContent>
            </Card>

            {/* Stage Distribution */}
            <Card className="border-0 shadow-xl">
              <CardHeader>
                <CardTitle className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <TrendingUp className="w-5 h-5 text-green-600" />
                    <span>Stage Distribution</span>
                  </div>
                  <Badge variant="secondary">By Stage</Badge>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {Object.entries(stageDistribution)
                    .sort(([,a], [,b]) => b - a)
                    .slice(0, 5)
                    .map(([stage, count], index) => {
                      const maxCount = Math.max(...Object.values(stageDistribution));
                      const percentage = maxCount > 0 ? (count / maxCount) * 100 : 0;
                      
                      return (
                        <div key={stage} className="flex items-center justify-between p-3 rounded-lg bg-slate-50 dark:bg-slate-800/50">
                          <div className="flex items-center space-x-3">
                            <div className="w-8 h-8 bg-blue-100 dark:bg-blue-900/20 rounded-lg flex items-center justify-center">
                              <span className="text-sm font-bold text-blue-600">#{index + 1}</span>
                            </div>
                            <div>
                              <p className="font-medium text-slate-900 dark:text-white">{stage}</p>
                              <p className="text-sm text-slate-600 dark:text-slate-400">{count} conversations</p>
                            </div>
                          </div>
                          <div className="text-right">
                            <p className="font-bold text-slate-900 dark:text-white">
                              {((count / stats.totalConversations) * 100).toFixed(1)}%
                            </p>
                            <div className="w-16 bg-slate-200 dark:bg-slate-700 rounded-full h-1.5 mt-1">
                              <div className="bg-green-500 h-1.5 rounded-full" style={{ width: `${percentage}%` }} />
                            </div>
                          </div>
                        </div>
                      );
                    })}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Additional Insights */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Card className="border-0 shadow-lg">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-slate-600 dark:text-slate-400 mb-1">
                      Conversations with Stages
                    </p>
                    <p className="text-lg font-bold text-white">{stats.conversationsWithStages}</p>
                    <p className="text-xs text-slate-400 mt-1">
                      {stats.totalConversations > 0 
                        ? `${((stats.conversationsWithStages / stats.totalConversations) * 100).toFixed(1)}% of total`
                        : 'No data'
                      }
                    </p>
                  </div>
                  <div className="w-10 h-10 bg-yellow-100 dark:bg-yellow-900/20 rounded-lg flex items-center justify-center">
                    <Activity className="w-5 h-5 text-yellow-600" />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-0 shadow-lg">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-slate-600 dark:text-slate-400 mb-1">
                      Unique Niches
                    </p>
                    <p className="text-lg font-bold text-white">{stats.uniqueNiches}</p>
                    <p className="text-xs text-slate-400 mt-1">Different conversation topics</p>
                  </div>
                  <div className="w-10 h-10 bg-indigo-100 dark:bg-indigo-900/20 rounded-lg flex items-center justify-center">
                    <Zap className="w-5 h-5 text-indigo-600" />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-0 shadow-lg">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-slate-600 dark:text-slate-400 mb-1">
                      AI Success Rate
                    </p>
                    <p className="text-lg font-bold text-white">
                      {stats.totalConversations > 0 
                        ? `${(((stats.totalConversations - stats.humanTakeovers) / stats.totalConversations) * 100).toFixed(1)}%`
                        : '0%'
                      }
                    </p>
                    <p className="text-xs text-slate-400 mt-1">{stats.humanTakeovers} human takeovers</p>
                  </div>
                  <div className="w-10 h-10 bg-cyan-100 dark:bg-cyan-900/20 rounded-lg flex items-center justify-center">
                    <Globe className="w-5 h-5 text-cyan-600" />
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Data Table */}
          <Card className="border-0 shadow-lg">
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle className="flex items-center space-x-2">
                  <MessageSquare className="w-5 h-5" />
                  <span>AI WhatsApp Conversations</span>
                </CardTitle>
                <div className="flex items-center space-x-2">
                  <Button onClick={handleRefresh} variant="outline" size="sm">
                    <RefreshCw className="w-4 h-4 mr-2" />
                    Refresh
                  </Button>
                </div>
              </div>
              <div className="text-sm text-muted-foreground mt-2">
                Showing {startIndex + 1} to {Math.min(endIndex, filteredConversations.length)} of {filteredConversations.length} conversations
              </div>
            </CardHeader>
            <CardContent>
              {filteredConversations.length === 0 ? (
                <div className="text-center py-12">
                  <MessageSquare className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
                  <p className="text-muted-foreground">No conversations found for selected filters</p>
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
                        {paginatedConversations.map((conv, index) => (
                          <TableRow key={conv.id_prospect}>
                            <TableCell>{startIndex + index + 1}</TableCell>
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
                      Page {currentPage} of {totalPages}
                    </div>
                    <div className="flex items-center space-x-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setCurrentPage(currentPage - 1)}
                        disabled={currentPage === 1}
                      >
                        <ChevronLeft className="w-4 h-4" />
                        Previous
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setCurrentPage(currentPage + 1)}
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
        </>
      )}

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

export default Analytics;
