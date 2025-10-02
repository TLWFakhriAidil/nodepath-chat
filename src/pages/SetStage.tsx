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
import { useNavigate } from 'react-router-dom';
import { useDevice } from '@/contexts/DeviceContext';
import Swal from 'sweetalert2';

interface StageSetValue {
  stageSetValue_id: number;
  id_device: string;
  stage: string;  // Changed from number to string
  type_inputData: 'User Input' | 'Set';
  columnsData: string;
  inputHardCode: string | null;
}

export default function SetStage() {
  const { toast } = useToast();
  const navigate = useNavigate();
  const { has_devices, checkDeviceStatus, device_ids } = useDevice();
  const [stageValues, setStageValues] = useState<StageSetValue[]>([]);
  const [loading, setLoading] = useState(false);
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<StageSetValue | null>(null);
  const [hasCheckedDevices, setHasCheckedDevices] = useState(false);
  const [formData, setFormData] = useState({
    id_device: '',
    stage: '',
    type_inputData: 'User Input',
    columnsData: 'nama',
    inputHardCode: ''
  });

  useEffect(() => {
    // Check device status on mount
    const checkAndFetch = async () => {
      await checkDeviceStatus();
      setHasCheckedDevices(true);
      
      // If has devices, fetch stage values
      if (has_devices) {
        fetchStageValues();
      }
    };
    
    checkAndFetch();
  }, []);

  useEffect(() => {
    // Check if we need to redirect
    if (hasCheckedDevices && !has_devices) {
      toast({
        title: "Device Required",
        description: "Please add a device first to access this feature",
        variant: "destructive",
      });
      
      setTimeout(() => {
        navigate('/device-settings');
      }, 1500);
    } else if (hasCheckedDevices && has_devices) {
      fetchStageValues();
    }
  }, [hasCheckedDevices, has_devices, navigate, toast]);

  const fetchStageValues = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/stage-values', {
        credentials: 'include'
      });
      if (response.ok) {
        const data = await response.json();
        setStageValues(data);
      } else {
        console.error('Failed to fetch stage values');
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
    if (!formData.id_device || formData.id_device.trim() === '') {
      toast({
        title: "Error",
        description: "Please select a device",
        variant: "destructive"
      });
      return;
    }

    if (!formData.stage || formData.stage.trim() === '') {
      toast({
        title: "Error",
        description: "Stage is required",
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
        credentials: 'include',
        body: JSON.stringify({
          id_device: formData.id_device,
          stage: formData.stage.trim(),
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
        const error = await response.json();
        throw new Error(error.error || 'Failed to add stage value');
      }
    } catch (error: any) {
      console.error('Error adding stage value:', error);
      toast({
        title: "Error",
        description: error.message || "Failed to add stage value",
        variant: "destructive"
      });
    }
  };

  const handleEdit = (item: StageSetValue) => {
    setEditingItem(item);
    setFormData({
      id_device: item.id_device,
      stage: item.stage.toString(),
      type_inputData: item.type_inputData,
      columnsData: item.columnsData,
      inputHardCode: item.inputHardCode || ''
    });
    setIsEditDialogOpen(true);
  };

  const handleUpdate = async () => {
    if (!editingItem) return;

    if (!formData.stage || formData.stage.trim() === '') {
      toast({
        title: "Error",
        description: "Stage is required",
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
      const response = await fetch(`/api/stage-values/${editingItem.stageSetValue_id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          stage: formData.stage.trim(),  // Send as string
          type_inputData: formData.type_inputData,
          columnsData: formData.columnsData,
          inputHardCode: formData.type_inputData === 'Set' ? formData.inputHardCode : null
        })
      });

      if (response.ok) {
        toast({
          title: "Success",
          description: "Stage value updated successfully",
        });
        setIsEditDialogOpen(false);
        setEditingItem(null);
        resetForm();
        fetchStageValues();
      } else {
        const error = await response.json();
        throw new Error(error.error || 'Failed to update stage value');
      }
    } catch (error: any) {
      console.error('Error updating stage value:', error);
      toast({
        title: "Error",
        description: error.message || "Failed to update stage value",
        variant: "destructive"
      });
    }
  };

  const handleDelete = async (id: number) => {
    const result = await Swal.fire({
      title: 'Are you sure?',
      text: 'Do you want to delete this stage value?',
      icon: 'warning',
      showCancelButton: true,
      confirmButtonColor: '#3085d6',
      cancelButtonColor: '#d33',
      confirmButtonText: 'Yes, delete it!',
      cancelButtonText: 'Cancel'
    });

    if (!result.isConfirmed) return;
    
    try {
      const response = await fetch(`/api/stage-values/${id}`, {
        method: 'DELETE',
        credentials: 'include'
      });

      if (response.ok) {
        toast({
          title: "Success",
          description: "Stage value deleted successfully",
        });
        fetchStageValues();
        Swal.fire('Deleted!', 'The stage value has been deleted.', 'success');
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
      Swal.fire('Error!', 'Failed to delete stage value', 'error');
    }
  };

  const resetForm = () => {
    setFormData({
      id_device: '',
      stage: '',
      type_inputData: 'User Input',
      columnsData: 'nama',
      inputHardCode: ''
    });
  };

  // Show loading or redirect message if still checking or no devices
  if (!hasCheckedDevices) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <h2 className="text-xl font-semibold mb-2">Loading...</h2>
          <p className="text-muted-foreground">Checking device status...</p>
        </div>
      </div>
    );
  }

  if (hasCheckedDevices && !has_devices) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <h2 className="text-2xl font-bold mb-2">Device Required</h2>
          <p className="text-muted-foreground">Redirecting to Device Settings...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6">
      {/* Header with Add Button */}
      <Card className="mb-6">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Set Stage Management</CardTitle>
              <CardDescription className="mt-2">
                Configure stage values for your devices
              </CardDescription>
            </div>
            <div className="flex gap-2">
              <Button onClick={fetchStageValues} variant="outline" size="sm">
                <RefreshCw className="w-4 h-4 mr-2" />
                Refresh
              </Button>
              <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
                <DialogTrigger asChild>
                  <Button className="bg-blue-600 hover:bg-blue-700">
                    <Plus className="w-4 h-4 mr-2" />
                    Add Set Stage
                  </Button>
                </DialogTrigger>
                <DialogContent className="sm:max-w-[425px]">
                  <DialogHeader>
                    <DialogTitle>Add Stage Value</DialogTitle>
                    <DialogDescription>
                      Configure stage value settings for your device
                    </DialogDescription>
                  </DialogHeader>
                  <div className="grid gap-4 py-4">
                    <div className="grid grid-cols-4 items-center gap-4">
                      <Label htmlFor="device" className="text-right">
                        Device <span className="text-red-500">*</span>
                      </Label>
                      <Select
                        value={formData.id_device}
                        onValueChange={(value) => setFormData({ ...formData, id_device: value })}
                      >
                        <SelectTrigger className="col-span-3">
                          <SelectValue placeholder="Select a device" />
                        </SelectTrigger>
                        <SelectContent>
                          {device_ids.map((deviceId) => (
                            <SelectItem key={deviceId} value={deviceId}>
                              {deviceId}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="grid grid-cols-4 items-center gap-4">
                      <Label htmlFor="stage" className="text-right">
                        Stage
                      </Label>
                      <Input
                        id="stage"
                        type="text"
                        className="col-span-3"
                        value={formData.stage}
                        onChange={(e) => setFormData({ ...formData, stage: e.target.value })}
                        placeholder="Enter stage (e.g., S1, Stage1, 1)"
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
                          <SelectItem value="cara_bayaran">Cara Bayaran</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                  <DialogFooter>
                    <Button variant="outline" onClick={() => {
                      setIsAddDialogOpen(false);
                      resetForm();
                    }}>
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
        </CardHeader>
      </Card>

      {/* Edit Dialog */}
      <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Edit Stage Value</DialogTitle>
            <DialogDescription>
              Update stage value settings
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="edit-device" className="text-right">
                Device
              </Label>
              <Input
                id="edit-device"
                type="text"
                className="col-span-3"
                value={formData.id_device}
                disabled
                title="Device cannot be changed"
              />
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="edit-stage" className="text-right">
                Stage
              </Label>
              <Input
                id="edit-stage"
                type="text"
                className="col-span-3"
                value={formData.stage}
                onChange={(e) => setFormData({ ...formData, stage: e.target.value })}
                placeholder="Enter stage (e.g., S1, Stage1, 1)"
              />
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="edit-type" className="text-right">
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
                <Label htmlFor="edit-inputHardCode" className="text-right">
                  Input Hard Code
                </Label>
                <Input
                  id="edit-inputHardCode"
                  className="col-span-3"
                  value={formData.inputHardCode}
                  onChange={(e) => setFormData({ ...formData, inputHardCode: e.target.value })}
                  placeholder="Enter hardcoded value"
                />
              </div>
            )}
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="edit-column" className="text-right">
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
                  <SelectItem value="cara_bayaran">Cara Bayaran</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => {
              setIsEditDialogOpen(false);
              setEditingItem(null);
              resetForm();
            }}>
              Cancel
            </Button>
            <Button onClick={handleUpdate}>
              Update Stage Value
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Data Table */}
      <Card>
        <CardHeader>
          <CardTitle>Stage Values</CardTitle>
          <CardDescription>
            All configured stage values for your devices
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
                  <TableHead className="w-[80px]">ID</TableHead>
                  <TableHead>Device ID</TableHead>
                  <TableHead>Stage</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Input Hard Code</TableHead>
                  <TableHead>Column</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {stageValues.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center text-muted-foreground">
                      No stage values configured. Click "Add Set Stage" to create one.
                    </TableCell>
                  </TableRow>
                ) : (
                  stageValues.map((item) => (
                    <TableRow key={item.stageSetValue_id}>
                      <TableCell className="font-medium">{item.stageSetValue_id}</TableCell>
                      <TableCell>{item.id_device}</TableCell>
                      <TableCell>{item.stage}</TableCell>
                      <TableCell>
                        <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                          item.type_inputData === 'Set' 
                            ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300'
                            : 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
                        }`}>
                          {item.type_inputData}
                        </span>
                      </TableCell>
                      <TableCell>{item.inputHardCode || '-'}</TableCell>
                      <TableCell>{item.columnsData}</TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleEdit(item)}
                          >
                            <Edit className="w-4 h-4 text-blue-500" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleDelete(item.stageSetValue_id)}
                          >
                            <Trash2 className="w-4 h-4 text-red-500" />
                          </Button>
                        </div>
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
