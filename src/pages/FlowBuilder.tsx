import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { 
  Save, 
  Play, 
  Download, 
  Upload, 
  Settings, 
  Eye,
  Maximize2,
  Minimize2,
  Grid3X3,
  Layers
} from 'lucide-react';
import ChatbotBuilder from '@/components/ChatbotBuilder';
import { useNavigate } from 'react-router-dom';

const FlowBuilder = () => {
  const navigate = useNavigate();
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [showGrid, setShowGrid] = useState(true);
  const [showMinimap, setShowMinimap] = useState(true);

  const handleTestFlow = (flowId: string) => {
    navigate(`/test-chat?flowId=${flowId}`);
  };

  const toolbarActions = [
    {
      icon: Save,
      label: 'Save Flow',
      action: () => {},
      variant: 'default' as const,
      className: 'bg-green-600 hover:bg-green-700'
    },
    {
      icon: Play,
      label: 'Test Flow',
      action: () => navigate('/test-chat'),
      variant: 'default' as const,
      className: 'bg-blue-600 hover:bg-blue-700'
    },
    {
      icon: Download,
      label: 'Export',
      action: () => {},
      variant: 'outline' as const
    },
    {
      icon: Upload,
      label: 'Import',
      action: () => {},
      variant: 'outline' as const
    },
    {
      icon: Eye,
      label: 'Preview',
      action: () => {},
      variant: 'outline' as const
    }
  ];

  const viewControls = [
    {
      icon: isFullscreen ? Minimize2 : Maximize2,
      label: isFullscreen ? 'Exit Fullscreen' : 'Fullscreen',
      action: () => setIsFullscreen(!isFullscreen),
      active: isFullscreen
    },
    {
      icon: Grid3X3,
      label: 'Grid',
      action: () => setShowGrid(!showGrid),
      active: showGrid
    },
    {
      icon: Layers,
      label: 'Minimap',
      action: () => setShowMinimap(!showMinimap),
      active: showMinimap
    }
  ];

  return (
    <div className={`${isFullscreen ? 'fixed inset-0 z-50 bg-white dark:bg-slate-900' : 'space-y-6'}`}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white mb-2">
            Flow Builder
          </h1>
          <p className="text-slate-600 dark:text-slate-400">
            Design and build your conversational flows with our visual editor
          </p>
        </div>
        
        <div className="flex items-center space-x-2">
          <Badge variant="secondary" className="bg-green-100 text-green-700 dark:bg-green-900/20 dark:text-green-400">
            Auto-saved
          </Badge>
        </div>
      </div>

      {/* Toolbar */}
      <Card className="border-0 shadow-lg bg-white/50 dark:bg-slate-800/50 backdrop-blur-sm">
        <CardContent className="p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              {toolbarActions.map((action, index) => {
                const Icon = action.icon;
                return (
                  <Button
                    key={index}
                    variant={action.variant}
                    size="sm"
                    onClick={action.action}
                    className={action.className}
                  >
                    <Icon className="w-4 h-4 mr-2" />
                    {action.label}
                  </Button>
                );
              })}
            </div>
            
            <div className="flex items-center space-x-1">
              {viewControls.map((control, index) => {
                const Icon = control.icon;
                return (
                  <Button
                    key={index}
                    variant={control.active ? 'default' : 'ghost'}
                    size="sm"
                    onClick={control.action}
                    className={control.active ? 'bg-blue-600 hover:bg-blue-700 text-white' : ''}
                  >
                    <Icon className="w-4 h-4" />
                    <span className="sr-only">{control.label}</span>
                  </Button>
                );
              })}
              
              <div className="w-px h-6 bg-slate-200 dark:bg-slate-700 mx-2" />
              
              <Button variant="ghost" size="sm">
                <Settings className="w-4 h-4" />
                <span className="sr-only">Settings</span>
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Flow Builder */}
      <Card className={`border-0 shadow-xl overflow-hidden ${isFullscreen ? 'h-[calc(100vh-120px)]' : 'h-[calc(100vh-280px)]'}`}>
        <CardContent className="p-0 h-full">
          <ChatbotBuilder onTestFlow={handleTestFlow} />
        </CardContent>
      </Card>

      {/* Status Bar */}
      {!isFullscreen && (
        <div className="flex items-center justify-between text-sm text-slate-500 dark:text-slate-400 px-4 py-2 bg-slate-50 dark:bg-slate-800/50 rounded-lg">
          <div className="flex items-center space-x-4">
            <span>Nodes: 5</span>
            <span>Connections: 4</span>
            <span>Last saved: 2 minutes ago</span>
          </div>
          
          <div className="flex items-center space-x-2">
            <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
            <span>Connected</span>
          </div>
        </div>
      )}
    </div>
  );
};

export default FlowBuilder;