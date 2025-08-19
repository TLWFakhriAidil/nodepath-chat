import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
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
  Search, 
  RefreshCw, 
  ChevronLeft, 
  ChevronRight,
  MessageSquare,
  User,
  Calendar,
  Filter,
  Download
} from 'lucide-react';
import { format } from 'date-fns';

// AI WhatsApp conversation interface
interface AIWhatsappConversation {
  id_prospect: string;
  id_device: string;
  prospect_num: string;
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

const AIWhatsappDataTable = () => {
  const [conversations, setConversations] = useState<AIWhatsappConversation[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalRecords, setTotalRecords] = useState(0);
  const [pageSize, setPageSize] = useState(10);
  
  // Filter state
  const [deviceFilter, setDeviceFilter] = useState('all');
  const [stageFilter, setStageFilter] = useState('all');
  const [searchTerm, setSearchTerm] = useState('');
  
  // Available devices and stages for filters
  const [availableDevices, setAvailableDevices] = useState<string[]>([]);
  const [availableStages, setAvailableStages] = useState<string[]>([]);

  // Fetch AI WhatsApp data from the backend
  const fetchAIWhatsappData = async () => {
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
      
      const response = await fetch(`/api/ai/whatsapp/data?${params}`);
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      
      const data: AIWhatsappDataResponse = await response.json();
      
      setConversations(data.data);
      setCurrentPage(data.pagination.current_page);
      setTotalPages(data.pagination.total_pages);
      setTotalRecords(data.pagination.total_records);
      
      // Extract unique devices and stages for filters
      const devices = [...new Set(data.data.map(conv => conv.id_device).filter(Boolean))];
      const stages = [...new Set(data.data.map(conv => conv.stage).filter(Boolean))];
      
      setAvailableDevices(devices);
      setAvailableStages(stages);
      
    } catch (err) {
      console.error('Error fetching AI WhatsApp data:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch data');
    } finally {
      setLoading(false);
    }
  };

  // Load data on component mount and when filters change
  useEffect(() => {
    fetchAIWhatsappData();
  }, [currentPage, pageSize, deviceFilter, stageFilter, searchTerm]);

  // Handle search with debounce
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (currentPage !== 1) {
        setCurrentPage(1); // Reset to first page when searching
      } else {
        fetchAIWhatsappData();
      }
    }, 500);

    return () => clearTimeout(timeoutId);
  }, [searchTerm]);

  // Reset filters
  const resetFilters = () => {
    setDeviceFilter('all');
    setStageFilter('all');
    setSearchTerm('');
    setCurrentPage(1);
  };

  // Export data (placeholder)
  const exportData = () => {
    // TODO: Implement export functionality
    console.log('Export data functionality to be implemented');
  };

  // Format conversation data for display
  const formatConvLast = (convLast: any) => {
    if (!convLast) return 'No data';
    
    try {
      if (typeof convLast === 'string') {
        // Try to parse as JSON
        const parsed = JSON.parse(convLast);
        if (parsed.Response && Array.isArray(parsed.Response)) {
          return parsed.Response
            .filter((item: any) => item.type === 'text')
            .map((item: any) => item.content)
            .join(' ')
            .substring(0, 100) + '...';
        }
        return convLast.substring(0, 100) + '...';
      }
      return JSON.stringify(convLast).substring(0, 100) + '...';
    } catch {
      return String(convLast).substring(0, 100) + '...';
    }
  };

  // Get status badge color
  const getStatusBadgeColor = (human: number) => {
    return human === 1 ? 'bg-orange-100 text-orange-800' : 'bg-green-100 text-green-800';
  };

  // Get stage badge color
  const getStageBadgeColor = (stage: string) => {
    const colors = {
      'Problem Identification': 'bg-blue-100 text-blue-800',
      'Solution Presentation': 'bg-purple-100 text-purple-800',
      'Closing': 'bg-green-100 text-green-800',
      'Follow Up': 'bg-yellow-100 text-yellow-800',
    };
    return colors[stage as keyof typeof colors] || 'bg-gray-100 text-gray-800';
  };

  return (
    <Card className="w-full">
      <CardHeader>
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
          <div>
            <CardTitle className="flex items-center gap-2">
              <MessageSquare className="w-5 h-5" />
              AI WhatsApp Conversations
            </CardTitle>
            <p className="text-sm text-muted-foreground mt-1">
              View and manage all AI WhatsApp conversation records
            </p>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={exportData}
              className="flex items-center gap-2"
            >
              <Download className="w-4 h-4" />
              Export
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={fetchAIWhatsappData}
              disabled={loading}
              className="flex items-center gap-2"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          </div>
        </div>
      </CardHeader>
      
      <CardContent>
        {/* Filters */}
        <div className="flex flex-col sm:flex-row gap-4 mb-6">
          <div className="flex-1">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
              <Input
                placeholder="Search by phone number, niche, stage, or marketer..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>
          
          <Select value={deviceFilter} onValueChange={setDeviceFilter}>
            <SelectTrigger className="w-full sm:w-48">
              <SelectValue placeholder="Filter by Device" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Devices</SelectItem>
              {availableDevices.map((device) => (
                <SelectItem key={device} value={device}>
                  {device}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          
          <Select value={stageFilter} onValueChange={setStageFilter}>
            <SelectTrigger className="w-full sm:w-48">
              <SelectValue placeholder="Filter by Stage" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Stages</SelectItem>
              {availableStages.map((stage) => (
                <SelectItem key={stage} value={stage}>
                  {stage}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          
          <Button
            variant="outline"
            onClick={resetFilters}
            className="flex items-center gap-2"
          >
            <Filter className="w-4 h-4" />
            Reset
          </Button>
        </div>

        {/* Error Display */}
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4">
            <p className="font-medium">Error loading data:</p>
            <p className="text-sm">{error}</p>
          </div>
        )}

        {/* Loading State */}
        {loading && (
          <div className="flex items-center justify-center py-8">
            <RefreshCw className="w-6 h-6 animate-spin mr-2" />
            <span>Loading conversations...</span>
          </div>
        )}

        {/* Data Table */}
        {!loading && !error && (
          <>
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Phone Number</TableHead>
                    <TableHead>Device</TableHead>
                    <TableHead>Stage</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Niche</TableHead>
                    <TableHead>Marketer</TableHead>
                    <TableHead>Last Conversation</TableHead>
                    <TableHead>Updated</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {conversations.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={8} className="text-center py-8">
                        <div className="flex flex-col items-center gap-2">
                          <MessageSquare className="w-8 h-8 text-gray-400" />
                          <p className="text-gray-500">No conversations found</p>
                          <p className="text-sm text-gray-400">
                            Try adjusting your filters or search terms
                          </p>
                        </div>
                      </TableCell>
                    </TableRow>
                  ) : (
                    conversations.map((conversation) => (
                      <TableRow key={conversation.id_prospect}>
                        <TableCell className="font-medium">
                          {conversation.prospect_num || 'N/A'}
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">
                            {conversation.id_device}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge className={getStageBadgeColor(conversation.stage)}>
                            {conversation.stage || 'No Stage'}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge className={getStatusBadgeColor(conversation.human)}>
                            {conversation.human === 1 ? 'Human' : 'AI'}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {conversation.niche || 'N/A'}
                        </TableCell>
                        <TableCell>
                          {conversation.marketer || 'N/A'}
                        </TableCell>
                        <TableCell className="max-w-xs">
                          <div className="truncate" title={formatConvLast(conversation.conv_last)}>
                            {formatConvLast(conversation.conv_last)}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-1 text-sm text-gray-500">
                            <Calendar className="w-3 h-3" />
                            {format(new Date(conversation.updated_at), 'MMM dd, HH:mm')}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>

            {/* Pagination */}
            <div className="flex flex-col sm:flex-row justify-between items-center gap-4 mt-6">
              <div className="flex items-center gap-2 text-sm text-gray-600">
                <span>Showing</span>
                <Select value={pageSize.toString()} onValueChange={(value) => {
                  setPageSize(Number(value));
                  setCurrentPage(1);
                }}>
                  <SelectTrigger className="w-20">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="10">10</SelectItem>
                    <SelectItem value="25">25</SelectItem>
                    <SelectItem value="50">50</SelectItem>
                    <SelectItem value="100">100</SelectItem>
                  </SelectContent>
                </Select>
                <span>of {totalRecords} records</span>
              </div>
              
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                  disabled={currentPage === 1 || loading}
                  className="flex items-center gap-1"
                >
                  <ChevronLeft className="w-4 h-4" />
                  Previous
                </Button>
                
                <div className="flex items-center gap-1">
                  <span className="text-sm text-gray-600">
                    Page {currentPage} of {totalPages}
                  </span>
                </div>
                
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                  disabled={currentPage === totalPages || loading}
                  className="flex items-center gap-1"
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
  );
};

export default AIWhatsappDataTable;