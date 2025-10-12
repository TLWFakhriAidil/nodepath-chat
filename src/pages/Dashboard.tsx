import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { DatePickerWithRange } from '@/components/ui/date-picker';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { 
  Bot, 
  Workflow, 
  Upload, 
  BarChart3, 
  Plus,
  Play,
  Users,
  ArrowRight,
  Activity
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { getFlows } from '@/lib/localStorage';
import { ChatbotFlow } from '@/types/chatbot';
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { format } from 'date-fns';
import { DateRange } from 'react-day-picker';
import { useDevice } from '@/contexts/DeviceContext';

const Dashboard = () => {
  const navigate = useNavigate();
  const { device_ids } = useDevice();
  const [flows, setFlows] = useState<ChatbotFlow[]>([]);
  // Stats state removed since stats cards were removed

  // Helper function to sanitize malformed dates
  const sanitizeDate = (dateStr: string): Date => {
    try {
      // Fix malformed dates like "2025-10-01T00:00:00ZT00:00:00Z"
      let cleanDate = dateStr;
      
      // Remove duplicate timezone info (T00:00:00Z appearing twice)
      if (cleanDate.includes('T00:00:00ZT00:00:00Z')) {
        cleanDate = cleanDate.replace('T00:00:00ZT00:00:00Z', 'T00:00:00Z');
      }
      
      // Handle other potential malformed patterns
      cleanDate = cleanDate.replace(/Z.*Z$/, 'Z'); // Remove duplicate Z suffixes
      
      const date = new Date(cleanDate);
      
      // If date is invalid, try parsing just the date part
      if (isNaN(date.getTime())) {
        const datePart = cleanDate.split('T')[0];
        return new Date(datePart);
      }
      
      return date;
    } catch (error) {
      console.error('Error parsing date:', dateStr, error);
      return new Date(); // Fallback to current date
    }
  };

  // Chart state
  const [chartData, setChartData] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [dateRange, setDateRange] = useState<DateRange | undefined>({
    from: new Date(new Date().getFullYear(), new Date().getMonth(), 1),
    to: new Date()
  });
  const [deviceFilter, setDeviceFilter] = useState<string>('all');

  useEffect(() => {
    const loadFlows = async () => {
      const savedFlows = await getFlows();
      setFlows(savedFlows);
      // Stats calculation removed since stats cards were removed
    };
    loadFlows();
  }, []);

  // Fetch chart data
  useEffect(() => {
    const fetchChartData = async () => {
      if (!dateRange?.from || !dateRange?.to) return;
      
      setLoading(true);
      try {
        const params = new URLSearchParams({
          dateFrom: format(dateRange.from, 'yyyy-MM-dd'),
          dateTo: format(dateRange.to, 'yyyy-MM-dd'),
          deviceFilter: deviceFilter
        });

        const response = await fetch(`/api/dashboard/chart-data?${params}`, {
          credentials: 'include'
        });

        if (response.ok) {
          const data = await response.json();
          setChartData(data.data);
        }
      } catch (error) {
        console.error('Failed to fetch chart data:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchChartData();
  }, [dateRange, deviceFilter]);

  const quickActions = [
    {
      title: 'Create New Flow',
      description: 'Start building a new chatbot flow',
      icon: Plus,
      action: () => navigate('/flow-builder'),
      color: 'bg-gradient-to-r from-blue-500 to-purple-600',
      textColor: 'text-white'
    },
    {
      title: 'Test Existing Flow',
      description: 'Test and debug your flows',
      icon: Play,
      action: () => console.log('Test chat removed'),
      color: 'bg-gradient-to-r from-green-500 to-emerald-600',
      textColor: 'text-white'
    },
    {
      title: 'Upload Media',
      description: 'Manage your media assets',
      icon: Upload,
      action: () => navigate('/media'),
      color: 'bg-gradient-to-r from-orange-500 to-red-600',
      textColor: 'text-white'
    },
    {
      title: 'View Analytics',
      description: 'Check performance metrics',
      icon: BarChart3,
      action: () => navigate('/analytics'),
      color: 'bg-gradient-to-r from-purple-500 to-pink-600',
      textColor: 'text-white'
    }
  ];

  // Stats cards removed as requested by user

  return (
    <div className="space-y-8">
      {/* Stats grid removed as requested by user */}

      {/* Dashboard Charts with Filters */}
      <div className="space-y-6">
        {/* Filters */}
        <Card>
          <CardHeader>
            <CardTitle>Analytics Overview</CardTitle>
            <CardDescription>
              View combined data from AI WhatsApp and WasapBot databases
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-4">
              <div className="flex-1 min-w-[250px]">
                <label className="text-sm font-medium mb-2 block">Date Range</label>
                <DatePickerWithRange
                  from={dateRange?.from}
                  to={dateRange?.to}
                  onSelect={setDateRange}
                />
              </div>
              <div className="min-w-[200px]">
                <label className="text-sm font-medium mb-2 block">Device Filter</label>
                <Select value={deviceFilter} onValueChange={setDeviceFilter}>
                  <SelectTrigger>
                    <SelectValue placeholder="All Devices" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Devices</SelectItem>
                    {device_ids && device_ids.map((deviceId) => (
                      <SelectItem key={deviceId} value={deviceId}>
                        {deviceId}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Combined Line Chart */}
        {loading ? (
          <Card>
            <CardContent className="flex items-center justify-center h-96">
              <div className="text-center">
                <Activity className="w-8 h-8 animate-spin mx-auto mb-4" />
                <p className="text-muted-foreground">Loading chart data...</p>
              </div>
            </CardContent>
          </Card>
        ) : chartData ? (
          <>
            <Card>
              <CardHeader>
                <CardTitle>Combined Analytics</CardTitle>
                <CardDescription>
                  Combined data from both AI WhatsApp and WasapBot
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="h-80">
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={(() => {
                      // Generate continuous date range for complete x-axis
                      const generateContinuousDateRange = () => {
                        if (!dateRange?.from || !dateRange?.to) return [];
                        
                        const start = new Date(dateRange.from);
                        const end = new Date(dateRange.to);
                        const dateArray = [];
                        
                        // Generate all dates between start and end
                        for (let d = new Date(start); d <= end; d.setDate(d.getDate() + 1)) {
                          dateArray.push(new Date(d));
                        }
                        
                        return dateArray;
                      };
                      
                      // Get data from both databases
                      const aiData = chartData.ai_whatsapp?.daily_data || [];
                      const wasapData = chartData.wasapbot?.daily_data || [];
                      
                      // Create data maps for lookup
                      const aiDataMap = new Map<string, number>();
                      const wasapDataMap = new Map<string, number>();
                      
                      // Process AI WhatsApp data
                      aiData.forEach((day: any) => {
                        try {
                          const date = sanitizeDate(day.date);
                          const dateKey = format(date, 'yyyy-MM-dd');
                          aiDataMap.set(dateKey, day.conversations || 0);
                        } catch (error) {
                          console.error('Error processing AI data date:', day.date, error);
                        }
                      });
                      
                      // Process WasapBot data
                      wasapData.forEach((day: any) => {
                        try {
                          const date = sanitizeDate(day.date);
                          const dateKey = format(date, 'yyyy-MM-dd');
                          wasapDataMap.set(dateKey, day.prospects || 0);
                        } catch (error) {
                          console.error('Error processing WasapBot data date:', day.date, error);
                        }
                      });
                      
                      // Generate continuous date range data
                      const continuousData = generateContinuousDateRange().map(date => {
                        const dateKey = format(date, 'yyyy-MM-dd');
                        const displayDate = format(date, 'MMM dd');
                        
                        return {
                          date: displayDate,
                          dateKey: dateKey,
                          'AI WhatsApp': aiDataMap.get(dateKey) || 0,
                          'WasapBot': wasapDataMap.get(dateKey) || 0
                        };
                      });
                      
                      // Return continuous data or fallback to summary if no date range
                      if (continuousData.length === 0) {
                        return [{ 
                          date: 'No Date Range', 
                          'AI WhatsApp': chartData.ai_whatsapp?.summary?.total_conversations || 0,
                          'WasapBot': chartData.wasapbot?.totalProspects || 0
                        }];
                      }
                      
                      return continuousData;
                    })()}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="date" />
                      <YAxis />
                      <Tooltip />
                      <Legend />
                      <Line 
                        type="monotone" 
                        dataKey="AI WhatsApp" 
                        stroke="#3b82f6" 
                        strokeWidth={2}
                        dot={{ fill: '#3b82f6', r: 6 }}
                      />
                      <Line 
                        type="monotone" 
                        dataKey="WasapBot" 
                        stroke="#10b981" 
                        strokeWidth={2}
                        dot={{ fill: '#10b981', r: 6 }}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
                <div className="grid grid-cols-2 gap-4 mt-4">
                  <div className="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
                    <div className="flex items-center gap-2 mb-2">
                      <div className="w-3 h-3 bg-blue-600 rounded-full"></div>
                      <span className="font-semibold">AI WhatsApp</span>
                    </div>
                    <div className="text-2xl font-bold text-blue-600">
                      {chartData.ai_whatsapp?.summary?.total_conversations || 0}
                    </div>
                    <div className="text-sm text-muted-foreground">Total Conversations</div>
                  </div>
                  <div className="p-4 bg-green-50 dark:bg-green-900/20 rounded-lg">
                    <div className="flex items-center gap-2 mb-2">
                      <div className="w-3 h-3 bg-green-600 rounded-full"></div>
                      <span className="font-semibold">WasapBot</span>
                    </div>
                    <div className="text-2xl font-bold text-green-600">
                      {chartData.wasapbot?.totalProspects || 0}
                    </div>
                    <div className="text-sm text-muted-foreground">Total Prospects</div>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Separate Side-by-Side Charts */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* AI WhatsApp Chart */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <div className="w-3 h-3 bg-blue-600 rounded-full"></div>
                    AI WhatsApp Stage Distribution
                  </CardTitle>
                  <CardDescription>
                    Conversation stages from ai_whatsapp_nodepath
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="flex justify-between items-center">
                      <span className="text-sm text-muted-foreground">Total Conversations</span>
                      <span className="text-lg font-bold text-blue-600">
                        {chartData.ai_whatsapp?.summary?.total_conversations || 0}
                      </span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm text-muted-foreground">AI Active</span>
                      <span className="text-lg font-bold">
                        {chartData.ai_whatsapp?.summary?.ai_active || 0}
                      </span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm text-muted-foreground">Human Takeover</span>
                      <span className="text-lg font-bold">
                        {chartData.ai_whatsapp?.summary?.human_takeover || 0}
                      </span>
                    </div>
                  </div>
                  <div className="h-64 mt-4">
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart data={
                        chartData.ai_whatsapp?.stage_distribution?.map((stage: any) => ({
                          stage: stage.stage,
                          count: stage.count
                        })) || []
                      }>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis 
                          dataKey="stage" 
                          angle={-45}
                          textAnchor="end"
                          height={80}
                          interval={0}
                          tick={{ fontSize: 11 }}
                        />
                        <YAxis />
                        <Tooltip />
                        <Bar 
                          dataKey="count" 
                          fill="#3b82f6"
                          radius={[8, 8, 0, 0]}
                        />
                      </BarChart>
                    </ResponsiveContainer>
                  </div>
                </CardContent>
              </Card>

              {/* WasapBot Chart */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <div className="w-3 h-3 bg-green-600 rounded-full"></div>
                    WasapBot Stage Breakdown
                  </CardTitle>
                  <CardDescription>
                    Prospect stages from wasapBot_nodepath
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div className="flex justify-between items-center">
                      <span className="text-sm text-muted-foreground">Total Prospects</span>
                      <span className="text-lg font-bold text-green-600">
                        {chartData.wasapbot?.totalProspects || 0}
                      </span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm text-muted-foreground">Active Executions</span>
                      <span className="text-lg font-bold">
                        {chartData.wasapbot?.activeExecutions || 0}
                      </span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-sm text-muted-foreground">Completed</span>
                      <span className="text-lg font-bold">
                        {chartData.wasapbot?.completedExecutions || 0}
                      </span>
                    </div>
                  </div>
                  <div className="h-64 mt-4">
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart data={
                        chartData.wasapbot?.stageBreakdown 
                          ? Object.entries(chartData.wasapbot.stageBreakdown).map(([stage, count]) => ({
                              stage,
                              count: count as number
                            }))
                          : []
                      }>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis 
                          dataKey="stage"
                          angle={-45}
                          textAnchor="end"
                          height={80}
                          interval={0}
                          tick={{ fontSize: 11 }}
                        />
                        <YAxis />
                        <Tooltip />
                        <Bar 
                          dataKey="count" 
                          fill="#10b981"
                          radius={[8, 8, 0, 0]}
                        />
                      </BarChart>
                    </ResponsiveContainer>
                  </div>
                </CardContent>
              </Card>
            </div>
          </>
        ) : (
          <Card>
            <CardContent className="flex items-center justify-center h-96">
              <div className="text-center">
                <BarChart3 className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
                <h3 className="text-lg font-semibold mb-2">No Data Available</h3>
                <p className="text-muted-foreground">
                  Select a date range and device filter to view analytics
                </p>
              </div>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Recent Flows */}
      <div>
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-2xl font-bold text-slate-900 dark:text-white">Recent Flows</h2>
          <Button 
            variant="outline" 
            onClick={() => navigate('/flow-builder')}
            className="border-slate-200 dark:border-slate-700"
          >
            View All
            <ArrowRight className="w-4 h-4 ml-2" />
          </Button>
        </div>
        
        {flows.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {flows.slice(0, 6).map((flow) => (
              <Card key={flow.id} className="group cursor-pointer transition-all duration-300 hover:shadow-lg border-0 bg-white/50 dark:bg-slate-800/50 backdrop-blur-sm">
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-lg text-slate-900 dark:text-white">
                      {flow.name || 'Untitled Flow'}
                    </CardTitle>
                    <Badge variant="secondary" className="bg-blue-100 text-blue-700 dark:bg-blue-900/20 dark:text-blue-400">
                      {flow.nodes.length} nodes
                    </Badge>
                  </div>
                  <CardDescription className="text-slate-600 dark:text-slate-400">
                    Created {new Date(flow.createdAt).toLocaleDateString()}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2 text-sm text-slate-500 dark:text-slate-400">
                      <Users className="w-4 h-4" />
                      <span>{Math.floor(Math.random() * 100)} interactions</span>
                    </div>
                    <Button 
                      size="sm" 
                      variant="ghost"
                      onClick={() => navigate('/flow-builder')}
                      className="opacity-0 group-hover:opacity-100 transition-opacity"
                    >
                      Edit
                      <ArrowRight className="w-3 h-3 ml-1" />
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        ) : (
          <Card className="border-0 bg-white/50 dark:bg-slate-800/50 backdrop-blur-sm">
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Workflow className="w-16 h-16 text-slate-400 mb-4" />
              <h3 className="text-lg font-semibold text-slate-900 dark:text-white mb-2">
                No flows yet
              </h3>
              <p className="text-slate-600 dark:text-slate-400 text-center mb-6">
                Create your first chatbot flow to get started with building conversational experiences.
              </p>
              <Button 
                onClick={() => navigate('/flow-builder')}
                className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
              >
                <Plus className="w-4 h-4 mr-2" />
                Create Your First Flow
              </Button>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
};

export default Dashboard;