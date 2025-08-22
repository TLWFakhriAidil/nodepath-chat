import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useToast } from '@/hooks/use-toast';
import { getFlows, deleteFlow } from '@/lib/mysqlStorage';
import { ChatbotFlow } from '@/types/chatbot';
import {
  Edit,
  Trash2,
  Plus,
  RefreshCw,
  Calendar,
  Workflow,
  Play
} from 'lucide-react';
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

export default function FlowManager() {
  const navigate = useNavigate();
  const { toast } = useToast();
  const [flows, setFlows] = useState<ChatbotFlow[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleting, setDeleting] = useState<string | null>(null);

  useEffect(() => {
    loadFlows();
  }, []);

  const loadFlows = async () => {
    try {
      setLoading(true);
      const flowsData = await getFlows();
      setFlows(flowsData);
    } catch (error) {
      console.error('Error loading flows:', error);
      toast({
        title: "Error",
        description: "Failed to load flows. Please try again.",
        variant: "destructive"
      });
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (flowId: string) => {
    navigate(`/flow-builder?id=${flowId}`);
  };

  const handleDelete = async (flowId: string) => {
    try {
      setDeleting(flowId);
      await deleteFlow(flowId);
      setFlows(flows.filter(flow => flow.id !== flowId));
      toast({
        title: "Success",
        description: "Flow deleted successfully.",
        variant: "default"
      });
    } catch (error) {
      console.error('Error deleting flow:', error);
      toast({
        title: "Error",
        description: "Failed to delete flow. Please try again.",
        variant: "destructive"
      });
    } finally {
      setDeleting(null);
    }
  };

  // handleSimulation function removed - test functionality no longer available

  const handleCreateNew = () => {
    navigate('/flow-builder');
  };

  const formatDate = (dateString: string) => {
    try {
      return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      });
    } catch {
      return 'Unknown';
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto p-6">
        <div className="flex items-center justify-center h-64">
          <RefreshCw className="w-8 h-8 animate-spin text-muted-foreground" />
          <span className="ml-2 text-muted-foreground">Loading flows...</span>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Flow Manager</h1>
          <p className="text-muted-foreground mt-1">
            Manage your chatbot flows - edit, delete, or test them
          </p>
        </div>
        <div className="flex gap-2">
          <Button onClick={loadFlows} variant="outline" size="sm">
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
          <Button onClick={handleCreateNew}>
            <Plus className="w-4 h-4 mr-2" />
            New Flow
          </Button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Flows</CardTitle>
            <Workflow className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{flows.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Flows</CardTitle>
            <Play className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{flows.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Recent Updates</CardTitle>
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {flows.filter(flow => {
                const updatedAt = new Date(flow.updatedAt || flow.createdAt || '');
                const dayAgo = new Date(Date.now() - 24 * 60 * 60 * 1000);
                return updatedAt > dayAgo;
              }).length}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Flows Data Table */}
      <Card>
        <CardHeader>
          <CardTitle>All Flows</CardTitle>
          <CardDescription>
            Manage your chatbot flows - view, edit, test, or delete them
          </CardDescription>
        </CardHeader>
        <CardContent>
          {flows.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12">
              <Workflow className="w-12 h-12 text-muted-foreground mb-4" />
              <h3 className="text-lg font-semibold mb-2">No flows found</h3>
              <p className="text-muted-foreground text-center mb-4">
                Get started by creating your first chatbot flow
              </p>
              <Button onClick={handleCreateNew}>
                <Plus className="w-4 h-4 mr-2" />
                Create Your First Flow
              </Button>
            </div>
          ) : (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-16">NO</TableHead>
                    <TableHead>ID</TableHead>
                    <TableHead>ID DEVICE</TableHead>
                    <TableHead>NAME</TableHead>
                    <TableHead>NICHE</TableHead>
                    <TableHead>CREATED_AT</TableHead>
                    <TableHead>UPDATED_AT</TableHead>
                    <TableHead className="text-center">ACTION</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {flows.map((flow, index) => (
                    <TableRow key={flow.id}>
                      <TableCell className="font-medium">{index + 1}</TableCell>
                      <TableCell className="font-mono text-xs">
                        <div className="max-w-32 truncate" title={flow.id}>
                          {flow.id}
                        </div>
                      </TableCell>
                      <TableCell>
                         {flow.selectedDeviceId || flow.id_device || flow.idDevice || 'Not set'}
                       </TableCell>
                      <TableCell className="font-medium">{flow.name}</TableCell>
                      <TableCell>
                        {flow.niche ? (
                          <Badge variant="outline">{flow.niche}</Badge>
                        ) : (
                          <span className="text-muted-foreground">Not set</span>
                        )}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {formatDate(flow.createdAt || '')}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {formatDate(flow.updatedAt || flow.createdAt || '')}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center justify-center gap-2">
                          {/* Test button removed */}
                          <Button
                            onClick={() => handleEdit(flow.id)}
                            variant="outline"
                            size="sm"
                            className="h-8 px-3"
                          >
                            <Edit className="w-3 h-3 mr-1" />
                            Edit
                          </Button>
                          <AlertDialog>
                            <AlertDialogTrigger asChild>
                              <Button
                                variant="outline"
                                size="sm"
                                className="h-8 px-3 text-destructive hover:text-destructive hover:bg-destructive/10"
                                disabled={deleting === flow.id}
                              >
                                <Trash2 className="w-3 h-3 mr-1" />
                                Delete
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
                                  onClick={() => handleDelete(flow.id)}
                                  className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                                >
                                  Delete
                                </AlertDialogAction>
                              </AlertDialogFooter>
                            </AlertDialogContent>
                          </AlertDialog>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}