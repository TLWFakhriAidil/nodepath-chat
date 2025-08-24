import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { 
  Bot, 
  Workflow, 
  MessageSquare, 
  Upload, 
  BarChart3, 
  Plus,
  Play,
  Users,
  TrendingUp,
  Clock,
  Zap,
  ArrowRight,
  Activity
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { getFlows } from '@/lib/localStorage';
import { ChatbotFlow } from '@/types/chatbot';

const Dashboard = () => {
  const navigate = useNavigate();
  const [flows, setFlows] = useState<ChatbotFlow[]>([]);
  const [stats, setStats] = useState({
    totalFlows: 0,
    activeFlows: 0,
    totalMessages: 0,
    responseTime: '1.2s'
  });

  useEffect(() => {
    const loadFlows = async () => {
      const savedFlows = await getFlows();
      setFlows(savedFlows);
      setStats({
        totalFlows: savedFlows.length,
        activeFlows: savedFlows.filter(f => f.nodes.length > 1).length,
        totalMessages: Math.floor(Math.random() * 1000) + 500,
        responseTime: '1.2s'
      });
    };
    loadFlows();
  }, []);

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

  const statCards = [
    {
      title: 'Total Flows',
      value: stats.totalFlows,
      icon: Workflow,
      change: '+12%',
      changeType: 'positive' as const,
      color: 'text-blue-600'
    },
    {
      title: 'Active Flows',
      value: stats.activeFlows,
      icon: Activity,
      change: '+8%',
      changeType: 'positive' as const,
      color: 'text-green-600'
    },
    {
      title: 'Messages Sent',
      value: stats.totalMessages.toLocaleString(),
      icon: MessageSquare,
      change: '+23%',
      changeType: 'positive' as const,
      color: 'text-purple-600'
    },
    {
      title: 'Avg Response Time',
      value: stats.responseTime,
      icon: Clock,
      change: '-5%',
      changeType: 'positive' as const,
      color: 'text-orange-600'
    }
  ];

  return (
    <div className="space-y-8">
      {/* Welcome Section */}
      <div className="bg-gradient-to-r from-blue-600 via-purple-600 to-indigo-600 rounded-2xl p-8 text-white relative overflow-hidden">
        <div className="absolute inset-0 bg-black/10" />
        <div className="relative z-10">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold mb-2">Welcome back!</h1>
              <p className="text-blue-100 text-lg">
                Ready to build amazing conversational experiences?
              </p>
            </div>
            <div className="hidden md:block">
              <Bot className="w-24 h-24 text-white/20" />
            </div>
          </div>
          
          <div className="mt-6 flex flex-wrap gap-4">
            <Button 
              onClick={() => navigate('/flow-builder')}
              className="bg-white/20 hover:bg-white/30 text-white border-white/30"
              size="lg"
            >
              <Plus className="w-4 h-4 mr-2" />
              Create New Flow
            </Button>
            <Button 
              onClick={() => navigate('/analytics')}
              variant="outline"
              className="border-white/30 text-white hover:bg-white/10"
              size="lg"
            >
              <BarChart3 className="w-4 h-4 mr-2" />
              View Analytics
            </Button>
          </div>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat, index) => {
          const Icon = stat.icon;
          return (
            <Card key={index} className="relative overflow-hidden border-0 shadow-lg bg-white/50 dark:bg-slate-800/50 backdrop-blur-sm">
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-sm font-medium text-slate-600 dark:text-slate-300">
                    {stat.title}
                  </CardTitle>
                  <Icon className={`w-4 h-4 ${stat.color}`} />
                </div>
              </CardHeader>
              <CardContent>
                <div className="flex items-center justify-between">
                  <div className="text-2xl font-bold text-slate-900 dark:text-white">
                    {stat.value}
                  </div>
                  <Badge 
                    variant="secondary" 
                    className={`${
                      stat.changeType === 'positive' 
                        ? 'bg-green-100 text-green-700 dark:bg-green-900/20 dark:text-green-400' 
                        : 'bg-red-100 text-red-700 dark:bg-red-900/20 dark:text-red-400'
                    }`}
                  >
                    <TrendingUp className="w-3 h-3 mr-1" />
                    {stat.change}
                  </Badge>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Quick Actions */}
      <div>
        <h2 className="text-2xl font-bold text-slate-900 dark:text-white mb-6">Quick Actions</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {quickActions.map((action, index) => {
            const Icon = action.icon;
            return (
              <Card 
                key={index} 
                className="group cursor-pointer transition-all duration-300 hover:scale-105 hover:shadow-xl border-0 overflow-hidden"
                onClick={action.action}
              >
                <CardContent className="p-0">
                  <div className={`${action.color} p-6 ${action.textColor}`}>
                    <Icon className="w-8 h-8 mb-4" />
                    <h3 className="font-semibold text-lg mb-2">{action.title}</h3>
                    <p className="text-sm opacity-90 mb-4">{action.description}</p>
                    <div className="flex items-center text-sm font-medium">
                      Get Started
                      <ArrowRight className="w-4 h-4 ml-2 group-hover:translate-x-1 transition-transform" />
                    </div>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
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