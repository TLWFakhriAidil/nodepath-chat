import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
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

const Analytics = () => {
  const [timeRange, setTimeRange] = useState('7d');
  const [refreshing, setRefreshing] = useState(false);

  // Mock analytics data
  const stats = {
    totalConversations: 1247,
    totalUsers: 892,
    avgResponseTime: 1.2,
    successRate: 94.5,
    conversationsChange: 12.5,
    usersChange: 8.3,
    responseTimeChange: -5.2,
    successRateChange: 2.1
  };

  const conversationData = [
    { date: '2024-01-08', conversations: 45, users: 32 },
    { date: '2024-01-09', conversations: 52, users: 38 },
    { date: '2024-01-10', conversations: 48, users: 35 },
    { date: '2024-01-11', conversations: 61, users: 42 },
    { date: '2024-01-12', conversations: 55, users: 39 },
    { date: '2024-01-13', conversations: 67, users: 48 },
    { date: '2024-01-14', conversations: 72, users: 51 }
  ];

  const topFlows = [
    { name: 'Customer Support', conversations: 324, success: 96.2 },
    { name: 'Product Inquiry', conversations: 287, success: 92.8 },
    { name: 'Order Status', conversations: 198, success: 98.1 },
    { name: 'Technical Help', conversations: 156, success: 89.4 },
    { name: 'General Info', conversations: 142, success: 94.7 }
  ];

  const handleRefresh = async () => {
    setRefreshing(true);
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 1000));
    setRefreshing(false);
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
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" size="sm">
                <Calendar className="w-4 h-4 mr-2" />
                {timeRange === '7d' ? 'Last 7 days' : 
                 timeRange === '30d' ? 'Last 30 days' : 
                 timeRange === '90d' ? 'Last 90 days' : 'All time'}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuItem onClick={() => setTimeRange('7d')}>
                Last 7 days
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setTimeRange('30d')}>
                Last 30 days
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setTimeRange('90d')}>
                Last 90 days
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setTimeRange('all')}>
                All time
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
          
          <Button className="bg-blue-600 hover:bg-blue-700" size="sm">
            <Download className="w-4 h-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card className="border-0 shadow-lg">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-slate-600 dark:text-slate-400 mb-1">
                  Total Conversations
                </p>
                <p className="text-2xl font-bold text-slate-900 dark:text-white">
                  {stats.totalConversations.toLocaleString()}
                </p>
                <div className="flex items-center mt-2">
                  {getChangeIcon(stats.conversationsChange)}
                  <span className={`text-sm font-medium ml-1 ${getChangeColor(stats.conversationsChange)}`}>
                    {Math.abs(stats.conversationsChange)}%
                  </span>
                  <span className="text-xs text-slate-500 ml-1">vs last period</span>
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
                  Active Users
                </p>
                <p className="text-2xl font-bold text-slate-900 dark:text-white">
                  {stats.totalUsers.toLocaleString()}
                </p>
                <div className="flex items-center mt-2">
                  {getChangeIcon(stats.usersChange)}
                  <span className={`text-sm font-medium ml-1 ${getChangeColor(stats.usersChange)}`}>
                    {Math.abs(stats.usersChange)}%
                  </span>
                  <span className="text-xs text-slate-500 ml-1">vs last period</span>
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
                  Avg Response Time
                </p>
                <p className="text-2xl font-bold text-slate-900 dark:text-white">
                  {stats.avgResponseTime}s
                </p>
                <div className="flex items-center mt-2">
                  {getChangeIcon(stats.responseTimeChange)}
                  <span className={`text-sm font-medium ml-1 ${getChangeColor(stats.responseTimeChange)}`}>
                    {Math.abs(stats.responseTimeChange)}%
                  </span>
                  <span className="text-xs text-slate-500 ml-1">vs last period</span>
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
                  Success Rate
                </p>
                <p className="text-2xl font-bold text-slate-900 dark:text-white">
                  {stats.successRate}%
                </p>
                <div className="flex items-center mt-2">
                  {getChangeIcon(stats.successRateChange)}
                  <span className={`text-sm font-medium ml-1 ${getChangeColor(stats.successRateChange)}`}>
                    {Math.abs(stats.successRateChange)}%
                  </span>
                  <span className="text-xs text-slate-500 ml-1">vs last period</span>
                </div>
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
        {/* Conversation Trends */}
        <Card className="border-0 shadow-xl">
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <BarChart3 className="w-5 h-5 text-blue-600" />
                <span>Conversation Trends</span>
              </div>
              <Badge variant="secondary">Daily</Badge>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {conversationData.map((day, index) => {
                const maxConversations = Math.max(...conversationData.map(d => d.conversations));
                const conversationWidth = (day.conversations / maxConversations) * 100;
                const userWidth = (day.users / maxConversations) * 100;
                
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
                        <span className="text-slate-600 dark:text-slate-400">
                          {day.users} users
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
                      <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-1">
                        <div 
                          className="bg-green-500 h-1 rounded-full transition-all duration-300" 
                          style={{ width: `${userWidth}%` }}
                        />
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
            <div className="flex items-center justify-center space-x-6 mt-6 pt-4 border-t">
              <div className="flex items-center space-x-2">
                <div className="w-3 h-3 bg-blue-600 rounded-full" />
                <span className="text-sm text-slate-600 dark:text-slate-400">Conversations</span>
              </div>
              <div className="flex items-center space-x-2">
                <div className="w-3 h-3 bg-green-500 rounded-full" />
                <span className="text-sm text-slate-600 dark:text-slate-400">Users</span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Top Performing Flows */}
        <Card className="border-0 shadow-xl">
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <TrendingUp className="w-5 h-5 text-green-600" />
                <span>Top Performing Flows</span>
              </div>
              <Badge variant="secondary">Success Rate</Badge>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {topFlows.map((flow, index) => (
                <div key={flow.name} className="flex items-center justify-between p-3 rounded-lg bg-slate-50 dark:bg-slate-800/50">
                  <div className="flex items-center space-x-3">
                    <div className="w-8 h-8 bg-blue-100 dark:bg-blue-900/20 rounded-lg flex items-center justify-center">
                      <span className="text-sm font-bold text-blue-600">#{index + 1}</span>
                    </div>
                    <div>
                      <p className="font-medium text-slate-900 dark:text-white">{flow.name}</p>
                      <p className="text-sm text-slate-600 dark:text-slate-400">
                        {flow.conversations} conversations
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="font-bold text-slate-900 dark:text-white">{flow.success}%</p>
                    <div className="w-16 bg-slate-200 dark:bg-slate-700 rounded-full h-1.5 mt-1">
                      <div 
                        className="bg-green-500 h-1.5 rounded-full" 
                        style={{ width: `${flow.success}%` }}
                      />
                    </div>
                  </div>
                </div>
              ))}
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
                  Peak Hours
                </p>
                <p className="text-lg font-bold text-slate-900 dark:text-white">
                  2:00 PM - 4:00 PM
                </p>
                <p className="text-xs text-slate-500 mt-1">Most active period</p>
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
                  Avg Session Duration
                </p>
                <p className="text-lg font-bold text-slate-900 dark:text-white">
                  4m 32s
                </p>
                <p className="text-xs text-slate-500 mt-1">Per conversation</p>
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
                  Global Reach
                </p>
                <p className="text-lg font-bold text-slate-900 dark:text-white">
                  23 Countries
                </p>
                <p className="text-xs text-slate-500 mt-1">Active regions</p>
              </div>
              <div className="w-10 h-10 bg-cyan-100 dark:bg-cyan-900/20 rounded-lg flex items-center justify-center">
                <Globe className="w-5 h-5 text-cyan-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default Analytics;