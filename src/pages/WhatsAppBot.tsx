import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useDevice } from '@/contexts/DeviceContext';
import DeviceRequiredPopup from '@/components/DeviceRequiredPopup';
import { cn } from '@/lib/utils';
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
  Calendar
} from 'lucide-react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

// WasapBot data interface
interface WasapBotRecord {
  id_prospect: number;
  flow_reference?: string;
  execution_id?: string;
  execution_status?: string;
  flow_id?: string;
  current_node_id?: string;
  prospect_num?: string;
  niche?: string;
  instance?: string;
  peringkat_sekolah?: string;
  alamat?: string;
  nama?: string;
  pakej?: string;
  no_fon?: string;
  cara_bayaran?: string;
  tarikh_gaji?: string;
  stage?: string;
  status?: string;
  umur?: string;
  kerja?: string;
  sijil?: string;
  date_last?: string;
}

/**
 * WhatsApp Bot component for managing WasapBot Exama flow data
 * Uses wasapBot_nodepath database table
 */
const WhatsAppBot = () => {
  const { has_devices, device_ids } = useDevice();
  const [refreshing, setRefreshing] = useState(false);
  const [wasapBotData, setWasapBotData] = useState<WasapBotRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedStatus, setSelectedStatus] = useState<string>('all');
  const [selectedStage, setSelectedStage] = useState<string>('all');
  const [showDeviceRequiredPopup, setShowDeviceRequiredPopup] = useState(false);

  // Statistics
  const [stats, setStats] = useState({
    totalProspects: 0,
    activeExecutions: 0,
    completedExecutions: 0,
    uniqueSchools: 0,
    uniquePackages: 0,
    totalWithPhone: 0
  });

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
      
      const params = new URLSearchParams();
      
      // Add device filter
      if (device_ids && device_ids.length > 0) {
        params.append('deviceIds', device_ids.join(','));
      }
      
      if (searchTerm) {
        params.append('search', searchTerm);
      }
      
      if (selectedStatus !== 'all') {
        params.append('status', selectedStatus);
      }
      
      if (selectedStage !== 'all') {
        params.append('stage', selectedStage);
      }
      
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
      
      // Set the data
      setWasapBotData(data.records || []);
      
      // Calculate statistics
      const records = data.records || [];
      setStats({
        totalProspects: records.length,
        activeExecutions: records.filter((r: WasapBotRecord) => r.execution_status === 'active').length,
        completedExecutions: records.filter((r: WasapBotRecord) => r.execution_status === 'completed').length,
        uniqueSchools: new Set(records.map((r: WasapBotRecord) => r.peringkat_sekolah).filter(Boolean)).size,
        uniquePackages: new Set(records.map((r: WasapBotRecord) => r.pakej).filter(Boolean)).size,
        totalWithPhone: records.filter((r: WasapBotRecord) => r.no_fon).length
      });
      
    } catch (err) {
      console.error('WhatsAppBot: Error fetching data:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch WasapBot data');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  // Fetch data on component mount and when filters change
  useEffect(() => {
    fetchWasapBotData();
  }, [has_devices, device_ids, searchTerm, selectedStatus, selectedStage]);

  const handleRefresh = () => {
    setRefreshing(true);
    fetchWasapBotData();
  };

  const handleExport = () => {
    // TODO: Implement CSV export
    console.log('Export WasapBot data');
  };

  // Filter data based on search term
  const filteredData = wasapBotData.filter(record => {
    if (!searchTerm) return true;
    const searchLower = searchTerm.toLowerCase();
    return (
      record.nama?.toLowerCase().includes(searchLower) ||
      record.prospect_num?.toLowerCase().includes(searchLower) ||
      record.no_fon?.toLowerCase().includes(searchLower) ||
      record.alamat?.toLowerCase().includes(searchLower) ||
      record.peringkat_sekolah?.toLowerCase().includes(searchLower)
    );
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
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">WhatsApp Bot</h1>
          <p className="text-muted-foreground">
            Manage and monitor WasapBot Exama flow conversations
          </p>
        </div>
        <div className="flex gap-2">
          <Button 
            variant="outline" 
            onClick={handleExport}
            disabled={loading}
          >
            <Download className="w-4 h-4 mr-2" />
            Export
          </Button>
          <Button 
            onClick={handleRefresh}
            disabled={loading || refreshing}
          >
            <RefreshCw className={cn("w-4 h-4 mr-2", refreshing && "animate-spin")} />
            Refresh
          </Button>
        </div>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Prospects
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalProspects}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Active Flows
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{stats.activeExecutions}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Completed
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-600">{stats.completedExecutions}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Schools
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.uniqueSchools}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Packages
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.uniquePackages}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              With Phone
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalWithPhone}</div>
          </CardContent>
        </Card>
      </div>

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle>Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4">
            <Input
              placeholder="Search by name, phone, address..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="max-w-sm"
            />
            <Select value={selectedStatus} onValueChange={setSelectedStatus}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="Prospek">Prospek</SelectItem>
                <SelectItem value="Customer">Customer</SelectItem>
                <SelectItem value="Lead">Lead</SelectItem>
              </SelectContent>
            </Select>
            <Select value={selectedStage} onValueChange={setSelectedStage}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Stage" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Stages</SelectItem>
                <SelectItem value="welcome">Welcome</SelectItem>
                <SelectItem value="qualification">Qualification</SelectItem>
                <SelectItem value="presentation">Presentation</SelectItem>
                <SelectItem value="closing">Closing</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

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
                    <TableHead>Name</TableHead>
                    <TableHead>Phone</TableHead>
                    <TableHead>School</TableHead>
                    <TableHead>Package</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Stage</TableHead>
                    <TableHead>Payment</TableHead>
                    <TableHead>Last Updated</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredData.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={9} className="text-center py-8">
                        <MessageSquare className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
                        <p className="text-muted-foreground">No WasapBot data available</p>
                      </TableCell>
                    </TableRow>
                  ) : (
                    filteredData.map((record) => (
                      <TableRow key={record.id_prospect}>
                        <TableCell className="font-medium">
                          {record.nama || '-'}
                        </TableCell>
                        <TableCell>
                          {record.no_fon || record.prospect_num || '-'}
                        </TableCell>
                        <TableCell>
                          {record.peringkat_sekolah || '-'}
                        </TableCell>
                        <TableCell>
                          {record.pakej ? (
                            <Badge variant="outline">{record.pakej}</Badge>
                          ) : '-'}
                        </TableCell>
                        <TableCell>
                          <Badge variant={
                            record.status === 'Customer' ? 'default' : 
                            record.status === 'Lead' ? 'secondary' : 
                            'outline'
                          }>
                            {record.status || 'Prospek'}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          {record.stage || '-'}
                        </TableCell>
                        <TableCell>
                          {record.cara_bayaran || '-'}
                        </TableCell>
                        <TableCell>
                          {record.date_last || '-'}
                        </TableCell>
                        <TableCell>
                          <Button size="sm" variant="ghost">
                            View
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
