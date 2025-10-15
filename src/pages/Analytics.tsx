import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { DatePickerWithRange } from '@/components/ui/date-picker';
import AIWhatsappDataTable from '@/components/AIWhatsappDataTable';
import { useDevice } from '@/contexts/DeviceContext';
import DeviceRequiredPopup from '@/components/DeviceRequiredPopup';
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
  Globe
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { DateRange } from 'react-day-picker';
import { format } from 'date-fns';

// Analytics data interface
interface AnalyticsData {
  totalConversations: number;
  aiActiveConversations: number;
  humanTakeovers: number;
  uniqueDevices: number;
  uniqueNiches: number;
  conversationsWithStages: number;
  dailyBreakdown: Array<{
    date: string;
    conversations: number;
  }>;
  stageDistribution: Record<string, number>;
}

/**
 * Analytics component with device-based data filtering
 * Automatically filters analytics data by user's configured devices
 */
const Analytics = () => {
  const { has_devices, device_ids } = useDevice();
  const [timeRange, setTimeRange] = useState('current_month');
  const [refreshing, setRefreshing] = useState(false);
  const [analyticsData, setAnalyticsData] = useState<AnalyticsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dateRange, setDateRange] = useState<DateRange | undefined>({
    from: new Date(new Date().getFullYear(), new Date().getMonth(), 1), // Start of current month
    to: new Date() // Today
  });
  const [selectedDevice, setSelectedDevice] = useState<string>(''); // Empty string means all devices
  const [selectedStage, setSelectedStage] = useState<string>(''); // Empty string means all stages
  const [availableStages, setAvailableStages] = useState<string[]>([]);
  const [showDeviceRequiredPopup, setShowDeviceRequiredPopup] = useState(false);

  /**
   * Fetch analytics data from backend API
   * Automatically filters by user's device IDs from device context
   */
  const fetchAnalyticsData = async () => {
    // Remove device check - let backend handle it
    console.log('Analytics: Fetching data...');
    console.log('Analytics: Device IDs from context:', device_ids);
    
    try {
      setLoading(true);
      setError(null);
      
      const params = new URLSearchParams();
      
      if (dateRange?.from) {
        params.append('startDate', format(dateRange.from, 'yyyy-MM-dd'));
      }
      if (dateRange?.to) {
        params.append('endDate', format(dateRange.to, 'yyyy-MM-dd'));
      }
      
      // Use specific device if selected, otherwise use all user's devices
      if (selectedDevice) {
        params.append('idDevice', selectedDevice);
      } else if (device_ids && device_ids.length > 0) {
        // Send all device IDs as comma-separated string
        params.append('deviceIds', device_ids.join(','));
      }
      
      // Add stage filter if selected
      // Special case: "Welcome Message" means NULL/empty stage in database
      if (selectedStage) {
        if (selectedStage === 'Welcome Message') {
          params.append('stage', 'WELCOME_MESSAGE_NULL');
        } else {
          params.append('stage', selectedStage);
        }
      }
      
      const apiUrl = `/api/ai-whatsapp/ai/analytics?${params.toString()}`;
      console.log('Analytics: Making API call to:', apiUrl);
      console.log('Analytics: Device IDs from context:', device_ids);
      console.log('Analytics: Selected device:', selectedDevice);
      console.log('Analytics: Date range:', dateRange);
      
      const response = await fetch(apiUrl, {
        method: 'GET',
        credentials: 'include', // Use cookie-based authentication
        headers: {
          'Content-Type': 'application/json',
        },
      });
      
      console.log('Analytics: Response status:', response.status);
      console.log('Analytics: Response headers:', Object.fromEntries(response.headers.entries()));
      
      if (!response.ok) {
        const errorText = await response.text();
        console.error('Analytics: Error response body:', errorText);
        throw new Error(`HTTP error! status: ${response.status} - ${errorText}`);
      }
      
      const data = await response.json();
      console.log('Analytics: Received data:', data);
      setAnalyticsData(data);
    } catch (err) {
      console.error('Error fetching analytics data:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch analytics data');
    } finally {
      setLoading(false);
    }
  };
  
  // Update available stages from analytics data
  useEffect(() => {
    if (analyticsData?.stageDistribution) {
      const stages = Object.keys(analyticsData.stageDistribution).filter(Boolean);
      setAvailableStages(stages);
    }
  }, [analyticsData]);
  
  // Load data on component mount and when filters change
  useEffect(() => {
    fetchAnalyticsData();
  }, [dateRange, selectedDevice, selectedStage]);

  // Listen for refresh events from AIWhatsappDataTable
  useEffect(() => {
    const handleRefresh = () => {
      fetchAnalyticsData();
    };
    
    window.addEventListener('refreshAnalytics', handleRefresh);
    return () => window.removeEventListener('refreshAnalytics', handleRefresh);
  }, []);

  const handleRefresh = async () => {
    setRefreshing(true);
    await fetchAnalyticsData();
    setRefreshing(false);
  };
  
  const handleDateRangeChange = (range: DateRange | undefined) => {
    setDateRange(range);
    setTimeRange('custom');
  };
  
  const handleTimeRangeChange = (range: string) => {
    setTimeRange(range);
    const today = new Date();
    let from: Date;
    
    switch (range) {
      case '7d':
        from = new Date(today.getTime() - 7 * 24 * 60 * 60 * 1000);
        break;
      case '30d':
        from = new Date(today.getTime() - 30 * 24 * 60 * 60 * 1000);
        break;
      case '90d':
        from = new Date(today.getTime() - 90 * 24 * 60 * 60 * 1000);
        break;
      case 'current_month':
        from = new Date(today.getFullYear(), today.getMonth(), 1);
        break;
      default:
        from = new Date(today.getTime() - 30 * 24 * 60 * 60 * 1000);
    }
    
    setDateRange({ from, to: today });
  };
  
  // Calculate stats from analytics data
  const stats = analyticsData ? {
    totalConversations: analyticsData.totalConversations,
    aiActiveConversations: analyticsData.aiActiveConversations,
    humanTakeovers: analyticsData.humanTakeovers,
    uniqueDevices: analyticsData.uniqueDevices,
    conversationsChange: 0, // TODO: Calculate change from previous period
    aiActiveChange: 0,
    humanTakeoverChange: 0,
    devicesChange: 0
  } : {
    totalConversations: 0,
    aiActiveConversations: 0,
    humanTakeovers: 0,
    uniqueDevices: 0,
    conversationsChange: 0,
    aiActiveChange: 0,
    humanTakeoverChange: 0,
    devicesChange: 0
  };

  const getChangeIcon = (change: number) => {
    return change >= 0 ? (
      <ArrowUpRight className="w-4 h-4 text-green-600" />
    ) : (
      <ArrowDownRight className="w-4 h-4 text-red-600" />
    );
  };

  const getChangeColor = (change: number) => {
    return change >= 0 ? 'text-green-600' : 'text-red-600';
  };

  return (
    <div className="space-y-6">
      {/* Header */}
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
          <DatePickerWithRange
            from={dateRange?.from}
            to={dateRange?.to}
            onSelect={handleDateRangeChange}
            className="w-auto"
          />
          
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm">
                <Calendar className="w-4 h-4 mr-2" />
                {timeRange === '7d' ? 'Last 7 days' : 
                 timeRange === '30d' ? 'Last 30 days' : 
                 timeRange === '90d' ? 'Last 90 days' : 
                 timeRange === 'current_month' ? 'Current Month' : 'Custom Range'}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuItem onClick={() => handleTimeRangeChange('7d')}>
                Last 7 days
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => handleTimeRangeChange('30d')}>
                Last 30 days
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => handleTimeRangeChange('90d')}>
                Last 90 days
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => handleTimeRangeChange('current_month')}>
                Current Month
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          
          <Button 
            variant="outline" 
            size="sm" 
            onClick={handleRefresh}
            disabled={refreshing}
          >
            <RefreshCw className={`w-4 h-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
          
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm" disabled={!has_devices}>
                <Filter className="w-4 h-4 mr-2" />
                {selectedDevice ? `Device: ${selectedDevice}` : 'All Devices'}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuItem onClick={() => setSelectedDevice('')}>
                All Devices
              </DropdownMenuItem>
              {/* Dynamically render user's devices */}
              {device_ids && device_ids.map((deviceId) => (
                <DropdownMenuItem 
                  key={deviceId} 
                  onClick={() => setSelectedDevice(deviceId)}
                >
                  {deviceId}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
          
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm" disabled={!has_devices}>
                <Filter className="w-4 h-4 mr-2" />
                {selectedStage ? `Stage: ${selectedStage}` : 'All Stages'}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuItem onClick={() => setSelectedStage('')}>
                All Stages
              </DropdownMenuItem>
              {availableStages && availableStages.map((stage) => (
                <DropdownMenuItem 
                  key={stage} 
                  onClick={() => setSelectedStage(stage)}
                >
                  {stage || 'Welcome Message'}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
          
          <Button className="bg-blue-600 hover:bg-blue-700" size="sm">
            <Download className="w-4 h-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      {/* Loading and Error States */}
      {loading && (
        <div className="flex items-center justify-center py-8">
          <RefreshCw className="w-6 h-6 animate-spin mr-2" />
          <span>Loading analytics data...</span>
        </div>
      )}
      
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-red-800">Error: {error}</p>
          <Button onClick={handleRefresh} className="mt-2" size="sm">
            Try Again
          </Button>
        </div>
      )}

      {/* Key Metrics */}
      {!loading && !error && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="border-0 shadow-lg">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-slate-600 dark:text-slate-400 mb-1">
                    Total Conversations
                  </p>
                  <p className="text-2xl font-bold text-white">
                    {stats.totalConversations.toLocaleString()}
                  </p>
                  <div className="flex items-center mt-2">
                    {getChangeIcon(stats.conversationsChange)}
                    <span className={`text-sm font-medium ml-1 ${getChangeColor(stats.conversationsChange)}`}>
                      {Math.abs(stats.conversationsChange)}%
                    </span>
                    <span className="text-xs text-slate-400 ml-1">vs last period</span>
                  </div>
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
                  <p className="text-2xl font-bold text-white">
                    {stats.aiActiveConversations.toLocaleString()}
                  </p>
                  <div className="flex items-center mt-2">
                    {getChangeIcon(stats.aiActiveChange)}
                    <span className={`text-sm font-medium ml-1 ${getChangeColor(stats.aiActiveChange)}`}>
                      {Math.abs(stats.aiActiveChange)}%
                    </span>
                    <span className="text-xs text-slate-400 ml-1">vs last period</span>
                  </div>
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
                  <p className="text-2xl font-bold text-white">
                    {stats.humanTakeovers.toLocaleString()}
                  </p>
                  <div className="flex items-center mt-2">
                    {getChangeIcon(stats.humanTakeoverChange)}
                    <span className={`text-sm font-medium ml-1 ${getChangeColor(stats.humanTakeoverChange)}`}>
                      {Math.abs(stats.humanTakeoverChange)}%
                    </span>
                    <span className="text-xs text-slate-400 ml-1">vs last period</span>
                  </div>
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
                  <p className="text-2xl font-bold text-white">
                    {stats.uniqueDevices.toLocaleString()}
                  </p>
                  <div className="flex items-center mt-2">
                    {getChangeIcon(stats.devicesChange)}
                    <span className={`text-sm font-medium ml-1 ${getChangeColor(stats.devicesChange)}`}>
                      {Math.abs(stats.devicesChange)}%
                    </span>
                    <span className="text-xs text-slate-400 ml-1">vs last period</span>
                  </div>
                </div>
                <div className="w-12 h-12 bg-purple-100 dark:bg-purple-900/20 rounded-lg flex items-center justify-center">
                  <Target className="w-6 h-6 text-purple-600" />
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Charts Row */}
      {!loading && !error && analyticsData && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Conversation Trends */}
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
                {analyticsData.dailyBreakdown && analyticsData.dailyBreakdown.length > 0 ? (
                  analyticsData.dailyBreakdown.map((day, index) => {
                    const maxConversations = Math.max(...analyticsData.dailyBreakdown.map(d => d.conversations));
                    const conversationWidth = maxConversations > 0 ? (day.conversations / maxConversations) * 100 : 0;
                    
                    return (
                      <div key={day.date} className="space-y-2">
                        <div className="flex items-center justify-between text-sm">
                          <span className="text-slate-600 dark:text-slate-400">
                            {new Date(day.date).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })}
                          </span>
                          <div className="flex items-center space-x-4">
                            <span className="text-slate-900 dark:text-white font-medium">
                              {day.conversations} conversations
                            </span>
                          </div>
                        </div>
                        <div className="space-y-1">
                          <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2">
                            <div 
                              className="bg-blue-600 h-2 rounded-full transition-all duration-300" 
                              style={{ width: `${conversationWidth}%` }}
                            />
                          </div>
                        </div>
                      </div>
                    );
                  })
                ) : (
                  <div className="text-center py-8 text-slate-500">
                    No daily data available for the selected period
                  </div>
                )}
              </div>
              <div className="flex items-center justify-center space-x-6 mt-6 pt-4 border-t">
                <div className="flex items-center space-x-2">
                  <div className="w-3 h-3 bg-blue-600 rounded-full" />
                  <span className="text-sm text-slate-600 dark:text-slate-400">Conversations</span>
                </div>
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
                {analyticsData.stageDistribution && Object.keys(analyticsData.stageDistribution).length > 0 ? (
                  Object.entries(analyticsData.stageDistribution)
                    .sort(([,a], [,b]) => b - a)
                    .slice(0, 5)
                    .map(([stage, count], index) => {
                      const maxCount = Math.max(...Object.values(analyticsData.stageDistribution));
                      const percentage = maxCount > 0 ? (count / maxCount) * 100 : 0;
                      
                      return (
                        <div key={stage} className="flex items-center justify-between p-3 rounded-lg bg-slate-50 dark:bg-slate-800/50">
                          <div className="flex items-center space-x-3">
                            <div className="w-8 h-8 bg-blue-100 dark:bg-blue-900/20 rounded-lg flex items-center justify-center">
                              <span className="text-sm font-bold text-blue-600">#{index + 1}</span>
                            </div>
                            <div>
                              <p className="font-medium text-slate-900 dark:text-white">
                                {!stage || stage === 'No Stage' ? 'Welcome Message' : stage}
                              </p>
                              <p className="text-sm text-slate-600 dark:text-slate-400">
                                {count} conversations
                              </p>
                            </div>
                          </div>
                          <div className="text-right">
                            <p className="font-bold text-slate-900 dark:text-white">
                              {((count / analyticsData.totalConversations) * 100).toFixed(1)}%
                            </p>
                            <div className="w-16 bg-slate-200 dark:bg-slate-700 rounded-full h-1.5 mt-1">
                              <div 
                                className="bg-green-500 h-1.5 rounded-full" 
                                style={{ width: `${percentage}%` }}
                              />
                            </div>
                          </div>
                        </div>
                      );
                    })
                ) : (
                  <div className="text-center py-8 text-slate-500">
                    No stage data available for the selected period
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Additional Insights */}
      {!loading && !error && analyticsData && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="border-0 shadow-lg">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-slate-600 dark:text-slate-400 mb-1">
                    Conversations with Stages
                  </p>
                  <p className="text-lg font-bold text-white">
                    {analyticsData.conversationsWithStages.toLocaleString()}
                  </p>
                  <p className="text-xs text-slate-400 mt-1">
                    {analyticsData.totalConversations > 0 
                      ? `${((analyticsData.conversationsWithStages / analyticsData.totalConversations) * 100).toFixed(1)}% of total`
                      : 'No data available'
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
                  <p className="text-lg font-bold text-white">
                    {analyticsData.uniqueNiches.toLocaleString()}
                  </p>
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
                    {analyticsData.totalConversations > 0 
                      ? `${(((analyticsData.totalConversations - analyticsData.humanTakeovers) / analyticsData.totalConversations) * 100).toFixed(1)}%`
                      : '0%'
                    }
                  </p>
                  <p className="text-xs text-slate-400 mt-1">
                    {analyticsData.humanTakeovers} human takeovers
                  </p>
                </div>
                <div className="w-10 h-10 bg-cyan-100 dark:bg-cyan-900/20 rounded-lg flex items-center justify-center">
                  <Globe className="w-5 h-5 text-cyan-600" />
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
      
      {/* AI WhatsApp Data Table */}
      <div className="mt-8">
        <AIWhatsappDataTable 
          selectedDevice={selectedDevice} 
          selectedStage={selectedStage}
          dateRange={dateRange}
        />
      </div>
      
      {/* Device Required Popup */}
      <DeviceRequiredPopup 
        open={showDeviceRequiredPopup} 
        onOpenChange={setShowDeviceRequiredPopup} 
      />
    </div>
  );
};

export default Analytics;