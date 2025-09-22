import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useToast } from '@/hooks/use-toast';
import { Trash2, Plus, RefreshCw, Edit } from 'lucide-react';
import { useDevice } from '@/contexts/DeviceContext';

interface StageSetValue {
  stageSetValue_id: number;
  id_device: string;
  stage: number;
  type_inputData: 'User Input' | 'Set';
  columnsData: string;
  inputHardCode: string | null;
}

export default function SetStage() {
  const { toast } = useToast();
  const { selected_device } = useDevice();
  const [stageValues, setStageValues] = useState<StageSetValue[]>([]);
  const [loading, setLoading] = useState(false);
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [formData, setFormData] = useState({
    stage: '',
    type_inputData: 'User Input',
    columnsData: 'nama',
    inputHardCode: ''
  });

  useEffect(() => {
    if (selected_device) {
      fetchStageValues();
    }
  }, [selected_device]);

  const fetchStageValues = async () => {
    if (!selected_device) return;
    
    setLoading(true);
    try {
      const response = await fetch(`/api/stage-values/${selected_device}`);
      if (response.ok) {
        const data = await response.json();
        setStageValues(data);
      }
    } catch (error) {
      console.error('Error fetching stage values:', error);
      toast({
        title: "Error",
        description: "Failed to fetch stage values",
        variant: "destructive"
      });
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async () => {
    if (!selected_device) {
      toast({
        title: "Error",
        description: "Please select a device first",
        variant: "destructive"
      });
      return;
    }

    if (!formData.stage || isNaN(Number(formData.stage))) {
      toast({
        title: "Error",
        description: "Stage must be a valid number",
        variant: "destructive"
      });
      return;
    }

    if (formData.type_inputData === 'Set' && !formData.inputHardCode) {
      toast({
        title: "Error",
        description: "Input Hard Code is required when type is 'Set'",
        variant: "destructive"
      });
      return;
    }

    try {
      const response = await fetch('/api/stage-values', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          id_device: selected_device,
          stage: Number(formData.stage),
          type_inputData: formData.type_inputData,
          columnsData: formData.columnsData,
          inputHardCode: formData.type_inputData === 'Set' ? formData.inputHardCode : null
        })
      });

      if (response.ok) {
        toast({
          title: "Success",
          description: "Stage value added successfully",
        });
        setIsAddDialogOpen(false);
        resetForm();
        fetchStageValues();
      } else {
        throw new Error('Failed to add stage value');
      }
    } catch (error) {
      console.error('Error adding stage value:', error);
      toast({
        title: "Error",
        description: "Failed to add stage value",
        variant: "destructive"
      });
    }
  };

  const handleDelete = async (id: number) => {
    try {
      const response = await fetch(`/api/stage-values/${id}`, {
        method: 'DELETE'
      });

      if (response.ok) {
        toast({
          title: "Success",
          description: "Stage value deleted successfully",
        });
        fetchStageValues();
      } else {
        throw new Error('Failed to delete stage value');
      }
    } catch (error) {
      console.error('Error deleting stage value:', error);
      toast({
        title: "Error",
        description: "Failed to delete stage value",
        variant: "destructive"
      });
    }
  };

  const resetForm = () => {
    setFormData({
      stage: '',
      type_inputData: 'User Input',
      columnsData: 'nama',
      inputHardCode: ''
    });
  };

  if (!selected_device) {
    return (
      <div className="container mx-auto p-6">
        <Card>
          <CardHeader>
            <CardTitle>Set Stage</CardTitle>
            <CardDescription>Please select a device to manage stage values</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Set Stage</h1>
          <p className="text-muted-foreground mt-1">
            Manage stage values for device: {selected_device}
          </p>
        </div>
        <div className="flex gap-2">
          <Button onClick={fetchStageValues} variant="outline" size="sm">
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
          <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="w-4 h-4 mr-2" />
                Add Set Stage
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[425px]">
              <DialogHeader>
                <DialogTitle>Add Stage Value</DialogTitle>
                <DialogDescription>
                  Configure stage value settings for the selected device
                </DialogDescription>
              </DialogHeader>
              <div className="grid gap-4 py-4">
                <div className="grid grid-cols-4 items-center gap-4">
                  <Label htmlFor="stage" className="text-right">
                    Stage
                  </Label>
                  <Input
                    id="stage"
                    type="number"
                    className="col-span-3"
                    value={formData.stage}
                    onChange={(e) => setFormData({ ...formData, stage: e.target.value })}
                    placeholder="Enter stage number"
                  />
                </div>
                <div className="grid grid-cols-4 items-center gap-4">
                  <Label htmlFor="type" className="text-right">
                    Type
                  </Label>
                  <Select
                    value={formData.type_inputData}
                    onValueChange={(value) => setFormData({ ...formData, type_inputData: value as 'User Input' | 'Set' })}
                  >
                    <SelectTrigger className="col-span-3">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="User Input">User Input</SelectItem>
                      <SelectItem value="Set">Set</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                {formData.type_inputData === 'Set' && (
                  <div className="grid grid-cols-4 items-center gap-4">
                    <Label htmlFor="inputHardCode" className="text-right">
                      Input Hard Code
                    </Label>
                    <Input
                      id="inputHardCode"
                      className="col-span-3"
                      value={formData.inputHardCode}
                      onChange={(e) => setFormData({ ...formData, inputHardCode: e.target.value })}
                      placeholder="Enter hardcoded value"
                    />
                  </div>
                )}
                <div className="grid grid-cols-4 items-center gap-4">
                  <Label htmlFor="column" className="text-right">
                    Column
                  </Label>
                  <Select
                    value={formData.columnsData}
                    onValueChange={(value) => setFormData({ ...formData, columnsData: value })}
                  >
                    <SelectTrigger className="col-span-3">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="nama">Nama</SelectItem>
                      <SelectItem value="alamat">Alamat</SelectItem>
                      <SelectItem value="pakej">Pakej</SelectItem>
                      <SelectItem value="no_fon">No Fon</SelectItem>
                      <SelectItem value="tarikh_gaji">Tarikh Gaji</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsAddDialogOpen(false)}>
                  Cancel
                </Button>
                <Button onClick={handleSubmit}>
                  Add Stage Value
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {/* Data Table */}
      <Card>
        <CardHeader>
          <CardTitle>Stage Values</CardTitle>
          <CardDescription>
            All configured stage values for the selected device
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center h-32">
              <RefreshCw className="w-6 h-6 animate-spin text-muted-foreground" />
              <span className="ml-2 text-muted-foreground">Loading...</span>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[80px]">Stage</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Column</TableHead>
                  <TableHead>Input Hard Code</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {stageValues.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} className="text-center text-muted-foreground">
                      No stage values configured
                    </TableCell>
                  </TableRow>
                ) : (
                  stageValues.map((item) => (
                    <TableRow key={item.stageSetValue_id}>
                      <TableCell className="font-medium">{item.stage}</TableCell>
                      <TableCell>
                        <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                          item.type_inputData === 'Set' 
                            ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300'
                            : 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
                        }`}>
                          {item.type_inputData}
                        </span>
                      </TableCell>
                      <TableCell>{item.columnsData}</TableCell>
                      <TableCell>
                        {item.inputHardCode || '-'}
                      </TableCell>
                      <TableCell className="text-right">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDelete(item.stageSetValue_id)}
                        >
                          <Trash2 className="w-4 h-4 text-red-500" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
