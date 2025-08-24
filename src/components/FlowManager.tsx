import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { 
  Play, 
  Trash2, 
  Download, 
  Plus, 
  MessageSquare,
  FileText,
  Clock,
  Edit
} from 'lucide-react';

// Ensure Play icon is properly imported
const PlayIcon = Play;
import { ChatbotFlow } from '@/types/chatbot';
import { getFlows, deleteFlow } from '@/lib/localStorage';
import { useToast } from '@/hooks/use-toast';
import FlowSelector from './FlowSelector';
import FlowPreview from './FlowPreview';

interface FlowManagerProps {
  onCreateNew: () => void;
  onTestFlow: (flowId: string) => void;
}

export default function FlowManager({ onCreateNew, onTestFlow }: FlowManagerProps) {
  const navigate = useNavigate();
  const [flows, setFlows] = useState<ChatbotFlow[]>([]);
  const [selectedFlowId, setSelectedFlowId] = useState<string | null>(null);
  const [selectedFlow, setSelectedFlow] = useState<ChatbotFlow | null>(null);
  const { toast } = useToast();

  useEffect(() => {
    loadFlows();
  }, []);

  // Reload flows when component becomes visible (when navigating back)
  useEffect(() => {
    const handleFocus = () => {
      console.log('Window focused, reloading flows...');
      loadFlows();
    };
    
    window.addEventListener('focus', handleFocus);
    return () => window.removeEventListener('focus', handleFocus);
  }, []);

  useEffect(() => {
    if (selectedFlowId) {
      const flow = flows.find(f => f.id === selectedFlowId);
      setSelectedFlow(flow || null);
    } else {
      setSelectedFlow(null);
    }
  }, [selectedFlowId, flows]);

  const loadFlows = async () => {
    console.log('Loading flows in FlowManager...')
    const savedFlows = await getFlows();
    console.log('Loaded flows:', savedFlows)
    setFlows(savedFlows);
    
    // If we had a selected flow and it still exists, keep it selected
    if (selectedFlowId && !savedFlows.find(f => f.id === selectedFlowId)) {
      setSelectedFlowId(null);
    }
  };

  const handleDeleteFlow = async (flowId: string) => {
    await deleteFlow(flowId);
    loadFlows();
    
    if (selectedFlowId === flowId) {
      setSelectedFlowId(null);
    }
    
    toast({
      title: "Flow deleted",
      description: "The flow has been deleted successfully"
    });
  };

  const exportFlow = (flow: ChatbotFlow) => {
    const dataStr = JSON.stringify(flow, null, 2);
    const dataBlob = new Blob([dataStr], { type: 'application/json' });
    const url = URL.createObjectURL(dataBlob);
    
    const link = document.createElement('a');
    link.href = url;
    link.download = `${flow.name.replace(/[^a-z0-9]/gi, '_').toLowerCase()}.json`;
    link.click();
    
    URL.revokeObjectURL(url);
    
    toast({
      title: "Flow exported",
      description: `"${flow.name}" has been exported as JSON`
    });
  };

  return (
    <div className="h-screen bg-background flex">
      {/* Sidebar - Flow List */}
      <Card className="w-80 bg-card border-border rounded-none border-r">
        <div className="p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-semibold text-foreground">My Flows</h2>
            <div className="flex gap-2">
              <Button 
                onClick={loadFlows} 
                size="sm" 
                variant="outline"
                className="text-xs"
              >
                Refresh
              </Button>
              <Button onClick={() => navigate('/')} size="sm">
                <Plus className="w-4 h-4 mr-2" />
                New Flow
              </Button>
            </div>
          </div>

          {/* Flow Selector */}
          <div className="mb-6">
            <FlowSelector 
              flows={flows}
              selectedFlowId={selectedFlowId}
              onFlowSelect={setSelectedFlowId}
            />
          </div>

          {/* Flow List */}
          <div className="space-y-3">
            <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
              Saved Flows ({flows.length})
            </h3>
            
            {flows.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <MessageSquare className="w-8 h-8 mx-auto mb-3 opacity-50" />
                <p className="text-sm">No flows created yet</p>
                <p className="text-xs">Create your first flow to get started</p>
              </div>
            ) : (
              <div className="space-y-2 max-h-96 overflow-y-auto">
                {flows.map((flow) => (
                  <Card 
                    key={flow.id} 
                    className={`p-3 cursor-pointer transition-colors hover:bg-muted/50 ${
                      selectedFlowId === flow.id ? 'bg-muted ring-1 ring-primary' : ''
                    }`}
                    onClick={() => setSelectedFlowId(flow.id)}
                  >
                    <div className="flex items-start justify-between mb-2">
                      <div className="flex-1 min-w-0">
                        <h4 className="font-medium text-sm truncate">{flow.name}</h4>
                        <p className="text-xs text-muted-foreground truncate">
                          {flow.description}
                        </p>
                      </div>
                    </div>
                    
                    <div className="flex items-center justify-between">
                      <div className="flex gap-1">
                        <Badge variant="secondary" className="text-xs">
                          {flow.nodes.length} nodes
                        </Badge>
                        <Badge variant="outline" className="text-xs">
                          {flow.edges.length} connections
                        </Badge>
                      </div>
                    </div>
                    
                    <div className="flex items-center justify-between mt-3 pt-2 border-t">
                      <span className="text-xs text-muted-foreground flex items-center">
                        <Clock className="w-3 h-3 mr-1" />
                        {new Date(flow.updatedAt).toLocaleDateString()}
                      </span>
                      
                      <div className="flex gap-1">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={(e) => {
                            e.stopPropagation();
                            navigate(`/?edit=${flow.id}`);
                          }}
                          className="h-6 px-2"
                        >
                          <Edit className="w-3 h-3" />
                        </Button>
                        
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={(e) => {
                            e.stopPropagation();
                            onTestFlow(flow.id);
                            navigate('/test');
                          }}
                          className="h-6 px-2"
                        >
                          <PlayIcon className="w-3 h-3" />
                        </Button>
                        
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={(e) => {
                            e.stopPropagation();
                            exportFlow(flow);
                          }}
                          className="h-6 px-2"
                        >
                          <Download className="w-3 h-3" />
                        </Button>
                        
                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={(e) => e.stopPropagation()}
                              className="h-6 px-2 hover:bg-destructive/20 hover:text-destructive"
                            >
                              <Trash2 className="w-3 h-3" />
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>Delete Flow</AlertDialogTitle>
                              <AlertDialogDescription>
                                Are you sure you want to delete "{flow.name}"? This action cannot be undone.
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>Cancel</AlertDialogCancel>
                              <AlertDialogAction
                                onClick={() => handleDeleteFlow(flow.id)}
                                className="bg-destructive hover:bg-destructive/90"
                              >
                                Delete
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </div>
                    </div>
                  </Card>
                ))}
              </div>
            )}
          </div>
        </div>
      </Card>

      {/* Main Content - Flow Preview */}
      <div className="flex-1 p-6">
        <FlowPreview flow={selectedFlow} />
      </div>
    </div>
  );
}
