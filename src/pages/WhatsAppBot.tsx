import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { DatePickerWithRange } from '@/components/ui/date-picker';
import { useDevice } from '@/contexts/DeviceContext';
import DeviceRequiredPopup from '@/components/DeviceRequiredPopup';
import { cn } from '@/lib/utils';
import Swal from 'sweetalert2';
import { format } from 'date-fns';
import { DateRange } from 'react-day-picker';
import { 
  MessageSquare, 
  Users, 
  Clock, 
  Target,
  Download,
  RefreshCw,
  Activity,
  Zap,
  Globe,
  School,
  MapPin,
  Package,
  Phone,
  CreditCard,
  Calendar,
  Layers,
  Filter
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
} from "@/components/ui/table";

// WasapBot data interface
interface WasapBotRecord {
  id_prospect: number;
  id_device?: string;
  nama?: string;
  prospect_num?: string;
  niche?: string;
  status?: string;
  stage?: string;
  alamat?: string;
  pakej?: string;
  cara_bayaran?: string;
  tarikh_gaji?: string;
  no_fon?: string;
  flow_reference?: string;
  execution_id?: string;
  execution_status?: string;
  flow_id?: string;
  current_node_id?: string;
  instance?: string;
  peringkat_sekolah?: string;
  umur?: string;
  kerja?: string;
  sijil?: string;
  date_start?: string;
  date_last?: string;
}

/**
 * Get first day of current month and today's date
 */
const getCurrentMonthDateRange = () => {
  const now = new Date();
  const firstDay = new Date(now.getFullYear(), now.getMonth(), 1);
  const today = now; // Today as the end date
  
  // Format as YYYY-MM-DD
  const formatDate = (date: Date) => {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  };
  
  return {
    startDate: formatDate(firstDay),
    endDate: formatDate(today)
  };
};

/**
 * WhatsApp Bot component for managing WasapBot Exama flow data
 * Uses wasapBot_nodepath database table
 */
const WhatsAppBot = () => {
  const { has_devices, device_ids } = useDevice();
  const [refreshing, setRefreshing] = useState(false);
  const [wasapBotData, setWasapBotData] = useState<WasapBotRecord[]>([]);
  const [allData, setAllData] = useState<WasapBotRecord[]>([]); // All fetched data
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedDevice, setSelectedDevice] = useState<string>('');
  const [selectedStatus, setSelectedStatus] = useState<string>('');
  const [selectedStage, setSelectedStage] = useState<string>('');
  const [showDeviceRequiredPopup, setShowDeviceRequiredPopup] = useState(false);
  
  // Date filter state using DateRange (matches Chatbot-AI)
  const [dateRange, setDateRange] = useState<DateRange | undefined>({
    from: new Date(new Date().getFullYear(), new Date().getMonth(), 1),
    to: new Date()
  });

  // Statistics
  const [stats, setStats] = useState({
    totalProspects: 0,
    activeFlows: 0,
    completed: 0,
    packages: 0,
    addresses: 0,
    names: 0,
    noPhone: 0,
    payments: 0
  });

  // Dynamic stage statistics
  const [stageStats, setStageStats] = useState<Record<string, number>>({});

  /**
   * Fetch WasapBot data from backend API
   */
  const fetchWasapBotData = async () => {
    if (!has_devices) {
      setShowDeviceRequiredPopup(true);
      setLoading(false);
      return;
    }
    
    try {
      setLoading(true);
      setError(null);
      
      const params = new URLSearchParams({
        page: '1',
        limit: '10000' // Get all records for client-side filtering
      });
      
      // Only send device_ids for initial fetch
      if (device_ids && device_ids.length > 0) {
        params.append('deviceIds', device_ids.join(','));
      }
      
      // NOTE: Date, stage, status, and search filtering done client-side (see filteredData below)
      // Backend doesn't handle these filters correctly
      
      const apiUrl = `/api/wasapbot/data?${params.toString()}`;
      console.log('WhatsAppBot: Making API call to:', apiUrl);
      
      const response = await fetch(apiUrl, {
        headers: {
          'Accept': 'application/json',
        },
      });
      
      if (!response.ok) {
        throw new Error(`Failed to fetch data: ${response.statusText}`);
      }
      
      const data = await response.json();
      console.log('WhatsAppBot: Received data:', data);
      
      // Set the data (stats will be calculated from filteredData, not here)
      setWasapBotData(data.records || []);
      
    } catch (err) {
      console.error('WhatsAppBot: Error fetching data:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch WasapBot data');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  // Fetch data on component mount only
  useEffect(() => {
    if (has_devices && device_ids.length > 0) {
      fetchWasapBotData();
    }
  }, []); // Only run once on mount

  // NOTE: Filters are now client-side only (no refetch needed)
  // Date, stage, status, device, and search are all filtered in filteredData below

  const handleRefresh = () => {
    setRefreshing(true);
    fetchWasapBotData();
  };

  const handleExport = () => {
    // Export only filtered data that's visible in the table
    if (filteredData.length === 0) {
      Swal.fire({
        icon: 'info',
        title: 'No Data',
        text: 'No data to export',
        confirmButtonColor: '#3b82f6'
      });
      return;
    }

    // Prepare CSV headers
    const headers = ['No', 'ID Device', 'Name', 'Prospect Number', 'Niche', 'Status', 'No Phone', 'Stage', 'Address', 'Package', 'Payment', 'Payday Date'];
    
    // Prepare CSV rows
    const csvData = filteredData.map((record, index) => {
      return [
        index + 1,
        record.id_device || '',
        record.nama || '',
        record.prospect_num || '',
        record.niche || '',
        record.status || '',
        record.no_fon || '',
        record.stage || 'No Stage',
        record.alamat || '',
        record.pakej || '',
        record.cara_bayaran || '',
        record.tarikh_gaji || ''
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
    const filename = `wasapbot_export_${dateStr}.csv`;
    
    link.setAttribute('href', url);
    link.setAttribute('download', filename);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    
    console.log(`Exported ${filteredData.length} records to ${filename}`);
  };

  const handleDelete = async (id: number) => {
    const result = await Swal.fire({
      title: 'Are you sure?',
      text: 'Do you want to delete this record?',
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
      const response = await fetch(`/api/wasapbot/data/${id}`, {
        method: 'DELETE',
        credentials: 'include',
      });
      
      if (response.ok) {
        // Refresh data after deletion
        fetchWasapBotData();
        Swal.fire('Deleted!', 'The record has been deleted.', 'success');
      } else {
        console.error('Failed to delete record');
        Swal.fire('Error!', 'Failed to delete record', 'error');
      }
    } catch (error) {
      console.error('Error deleting record:', error);
      Swal.fire('Error!', 'An error occurred while deleting the record', 'error');
    }
  };

  // Client-side filtering for ALL filters (date, device, stage, status, search)
  const filteredData = wasapBotData.filter(record => {
    // Date filtering (by date_start column)
    if (dateRange?.from && dateRange?.to && record.date_start) {
      const recordDate = new Date(record.date_start);
      const startDate = new Date(dateRange.from);
      const endDate = new Date(dateRange.to);
      
      startDate.setHours(0, 0, 0, 0);
      endDate.setHours(23, 59, 59, 999);
      
      if (recordDate < startDate || recordDate > endDate) {
        return false;
      }
    }
    
    // Device filtering
    if (selectedDevice && record.id_device !== selectedDevice) {
      return false;
    }
    
    // Stage filtering
    if (selectedStage) {
      const recordStage = (!record.stage || record.stage === '') ? 'No Stage' : record.stage;
      if (recordStage !== selectedStage) {
        return false;
      }
    }
    
    // Status filtering
    if (selectedStatus && selectedStatus !== 'all' && record.status !== selectedStatus) {
      return false;
    }
    
    // Search filtering
    if (searchTerm) {
      const searchLower = searchTerm.toLowerCase();
      const matches = (
        record.nama?.toLowerCase().includes(searchLower) ||
        record.prospect_num?.toLowerCase().includes(searchLower) ||
        record.no_fon?.toLowerCase().includes(searchLower) ||
        record.alamat?.toLowerCase().includes(searchLower) ||
        record.peringkat_sekolah?.toLowerCase().includes(searchLower)
      );
      if (!matches) {
        return false;
      }
    }
    
    return true;
  });

  // Calculate statistics from filtered data (not raw data)
  const calculatedStats = {
    totalProspects: filteredData.length,
    activeFlows: filteredData.filter((r) => r.current_node_id && r.current_node_id !== 'end').length,
    completed: filteredData.filter((r) => r.current_node_id === 'end').length,
    packages: filteredData.filter((r) => r.pakej && r.pakej !== '').length,
    addresses: filteredData.filter((r) => r.alamat && r.alamat !== '').length,
    names: filteredData.filter((r) => r.nama && r.nama !== '').length,
    noPhone: filteredData.filter((r) => r.no_fon && r.no_fon !== '').length,
    payments: filteredData.filter((r) => r.cara_bayaran && r.cara_bayaran !== '').length
  };

  // Get all available stages from ALL data (for dropdown)
  const allAvailableStages: Record<string, number> = {};
  wasapBotData.forEach((record) => {
    const stage = (!record.stage || record.stage === '') ? 'No Stage' : record.stage;
    allAvailableStages[stage] = (allAvailableStages[stage] || 0) + 1;
  });

  // Calculate dynamic stage statistics from filtered data (for display)
  const calculatedStageStats: Record<string, number> = {};
  filteredData.forEach((record) => {
    const stage = (!record.stage || record.stage === '') ? 'No Stage' : record.stage;
    calculatedStageStats[stage] = (calculatedStageStats[stage] || 0) + 1;
  });

  if (!has_devices) {
    return (
      <div className="space-y-6">
        <DeviceRequiredPopup 
          open={showDeviceRequiredPopup} 
          onOpenChange={setShowDeviceRequiredPopup} 
        />
        
        <div className="text-center py-12">
          <MessageSquare className="w-16 h-16 mx-auto text-muted-foreground mb-4" />
          <h2 className="text-2xl font-semibold mb-2">No Devices Configured</h2>
          <p className="text-muted-foreground mb-4">
            Please configure at least one device to view WhatsApp Bot data
          </p>
          <Button onClick={() => setShowDeviceRequiredPopup(true)}>
            Configure Device
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <div className="flex justify-between items-center mb-4">
          <div>
            <h1 className="text-3xl font-bold text-slate-900 dark:text-white mb-2">
              WhatsApp Bot
            </h1>
            <p className="text-slate-600 dark:text-slate-400">
              Manage and monitor WasapBot Exama flow conversations
            </p>
          </div>
        </div>

        {/* Filters Bar - Matches Chatbot-AI Layout */}
        <div className="flex items-center justify-between">
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
            
            <Button variant="outline" size="sm" onClick={handleRefresh} disabled={loading || refreshing}>
              <RefreshCw className={cn("w-4 h-4 mr-2", refreshing && "animate-spin")} />
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
                <DropdownMenuItem onClick={() => setSelectedStage('No Stage')}>
                  No Stage
                </DropdownMenuItem>
                {Object.keys(allAvailableStages).filter(stage => stage !== 'No Stage').map((stage) => (
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
      </div>

      {/* Statistics Cards - Fixed boxes first */}
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-8 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Total Prospects</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{calculatedStats.totalProspects}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Active Flows</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{calculatedStats.activeFlows}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Complete</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-600">{calculatedStats.completed}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Package</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{calculatedStats.packages}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Address</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{calculatedStats.addresses}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Name</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{calculatedStats.names}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">No Phone</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{calculatedStats.noPhone}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Payment</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{calculatedStats.payments}</div>
          </CardContent>
        </Card>
      </div>

      {/* Dynamic Stage Statistics Cards */}
      {Object.keys(calculatedStageStats).length > 0 && (
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
          {Object.entries(calculatedStageStats).map(([stage, count]) => (
            <Card key={stage} className="border-dashed">
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-1">
                  <Layers className="w-3 h-3" />
                  {stage}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-purple-600">{count}</div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Data Table */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <MessageSquare className="w-5 h-5" />
            WasapBot Conversations
          </CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="text-center py-8">
              <RefreshCw className="w-8 h-8 animate-spin mx-auto mb-4" />
              <p className="text-muted-foreground">Loading WasapBot data...</p>
            </div>
          ) : error ? (
            <div className="text-center py-8">
              <p className="text-red-600">{error}</p>
              <Button onClick={handleRefresh} className="mt-4">
                Try Again
              </Button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>No</TableHead>
                    <TableHead>Created_at</TableHead>
                    <TableHead>ID Device</TableHead>
                    <TableHead>Name</TableHead>
                    <TableHead>Prospect Number</TableHead>
                    <TableHead>Niche</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>No Phone</TableHead>
                    <TableHead>Stage</TableHead>
                    <TableHead>Address</TableHead>
                    <TableHead>Package</TableHead>
                    <TableHead>Payment</TableHead>
                    <TableHead>Payday Date</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredData.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={13} className="text-center py-8">
                        <MessageSquare className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
                        <p className="text-muted-foreground">No WasapBot data available</p>
                      </TableCell>
                    </TableRow>
                  ) : (
                    filteredData.map((record, index) => (
                      <TableRow key={record.id_prospect}>
                        <TableCell>{index + 1}</TableCell>
                        <TableCell>
                          {record.date_start ? format(new Date(record.date_start), 'dd-MM-yyyy') : '-'}
                        </TableCell>
                        <TableCell>{record.id_device || '-'}</TableCell>
                        <TableCell>{record.nama || '-'}</TableCell>
                        <TableCell>{record.prospect_num || '-'}</TableCell>
                        <TableCell>{record.niche || '-'}</TableCell>
                        <TableCell>
                          <Badge variant={
                            record.status === 'active' ? 'default' :
                            record.status === 'completed' ? 'secondary' :
                            'outline'
                          }>
                            {record.status || '-'}
                          </Badge>
                        </TableCell>
                        <TableCell>{record.no_fon || '-'}</TableCell>
                        <TableCell>
                          <Badge variant="outline">
                            {record.stage || 'No Stage'}
                          </Badge>
                        </TableCell>
                        <TableCell>{record.alamat || '-'}</TableCell>
                        <TableCell>{record.pakej || '-'}</TableCell>
                        <TableCell>{record.cara_bayaran || '-'}</TableCell>
                        <TableCell>{record.tarikh_gaji || '-'}</TableCell>
                        <TableCell>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleDelete(record.id_prospect)}
                          >
                            Delete
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default WhatsAppBot;
