import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Pencil, Trash2, RefreshCw, Phone, Mail, Eye } from 'lucide-react'
import { Lead } from '@/types/leads'
import { useLeads } from '@/hooks/useLeads'
import { format } from 'date-fns'
import { toast } from 'sonner'
import Swal from 'sweetalert2'

interface LeadTableProps {
  data: Lead[]
  loading: boolean
  onRefresh: () => void
}

export const LeadTable = ({ data, loading, onRefresh }: LeadTableProps) => {
  const { updateLead, deleteLead } = useLeads()
  const [searchTerm, setSearchTerm] = useState('')
  const [editingLead, setEditingLead] = useState<Lead | null>(null)
  const [viewingLead, setViewingLead] = useState<Lead | null>(null)

  const getStatusBadgeVariant = (status: string) => {
    switch (status) {
      case 'new': return 'default'
      case 'contacted': return 'secondary'
      case 'qualified': return 'outline'
      case 'converted': return 'default'
      case 'lost': return 'destructive'
      default: return 'default'
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'new': return 'bg-blue-500'
      case 'contacted': return 'bg-yellow-500'
      case 'qualified': return 'bg-purple-500'
      case 'converted': return 'bg-green-500'
      case 'lost': return 'bg-red-500'
      default: return 'bg-gray-500'
    }
  }

  const filteredData = data.filter(lead =>
    (lead.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
     lead.phone?.toLowerCase().includes(searchTerm.toLowerCase()) ||
     lead.email?.toLowerCase().includes(searchTerm.toLowerCase()) ||
     lead.source.toLowerCase().includes(searchTerm.toLowerCase()) ||
     lead.campaign_name?.toLowerCase().includes(searchTerm.toLowerCase()))
  )

  const handleEdit = async (lead: Lead) => {
    setEditingLead(lead)
  }

  const handleSaveEdit = async () => {
    if (!editingLead) return
    
    try {
      await updateLead(editingLead.id, {
        name: editingLead.name,
        phone: editingLead.phone,
        email: editingLead.email,
        status: editingLead.status,
        notes: editingLead.notes
      })
      setEditingLead(null)
      onRefresh()
    } catch (error) {
      // Error is handled in the hook
    }
  }

  const handleDelete = async (id: string) => {
    const result = await Swal.fire({
      title: 'Are you sure?',
      text: 'Do you want to delete this lead?',
      icon: 'warning',
      showCancelButton: true,
      confirmButtonColor: '#3085d6',
      cancelButtonColor: '#d33',
      confirmButtonText: 'Yes, delete it!',
      cancelButtonText: 'Cancel'
    });

    if (result.isConfirmed) {
      try {
        await deleteLead(id)
        onRefresh()
        Swal.fire('Deleted!', 'The lead has been deleted.', 'success');
      } catch (error) {
        // Error is handled in the hook
        Swal.fire('Error!', 'Failed to delete lead', 'error');
      }
    }
  }

  const handleView = (lead: Lead) => {
    setViewingLead(lead)
  }

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Lead Details</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-[400px] flex items-center justify-center">
            <div className="text-muted-foreground">Loading leads...</div>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Lead Details ({filteredData.length} leads)</CardTitle>
          <Button
            variant="outline"
            size="sm"
            onClick={onRefresh}
            className="flex items-center gap-2"
          >
            <RefreshCw className="h-4 w-4" />
            Refresh
          </Button>
        </div>
        <div className="flex items-center gap-4 mt-4">
          <Input
            placeholder="Search leads..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="max-w-sm"
          />
        </div>
      </CardHeader>
      <CardContent>
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Contact</TableHead>
                <TableHead>Source</TableHead>
                <TableHead>Campaign</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredData.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center text-muted-foreground">
                    No leads found
                  </TableCell>
                </TableRow>
              ) : (
                filteredData.map((lead) => (
                  <TableRow key={lead.id}>
                    <TableCell className="font-medium">
                      {lead.name || 'Anonymous'}
                    </TableCell>
                    <TableCell>
                      <div className="space-y-1">
                        {lead.phone && (
                          <div className="flex items-center gap-1 text-sm">
                            <Phone className="h-3 w-3" />
                            {lead.phone}
                          </div>
                        )}
                        {lead.email && (
                          <div className="flex items-center gap-1 text-sm">
                            <Mail className="h-3 w-3" />
                            {lead.email}
                          </div>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="capitalize">
                        {lead.source}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {lead.campaign_name || '-'}
                    </TableCell>
                    <TableCell>
                      <Badge 
                        variant={getStatusBadgeVariant(lead.status)}
                        className="capitalize"
                      >
                        <div className={`w-2 h-2 rounded-full ${getStatusColor(lead.status)} mr-1`} />
                        {lead.status}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {format(new Date(lead.created_at), 'MMM dd, yyyy')}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleView(lead)}
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleEdit(lead)}
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDelete(lead.id)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </CardContent>

      {/* Edit Dialog */}
      <Dialog open={!!editingLead} onOpenChange={() => setEditingLead(null)}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Edit Lead</DialogTitle>
          </DialogHeader>
          {editingLead && (
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  value={editingLead.name || ''}
                  onChange={(e) => setEditingLead({...editingLead, name: e.target.value})}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="phone">Phone</Label>
                <Input
                  id="phone"
                  value={editingLead.phone || ''}
                  onChange={(e) => setEditingLead({...editingLead, phone: e.target.value})}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  value={editingLead.email || ''}
                  onChange={(e) => setEditingLead({...editingLead, email: e.target.value})}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="status">Status</Label>
                <Select
                  value={editingLead.status}
                  onValueChange={(value: any) => setEditingLead({...editingLead, status: value})}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="new">New</SelectItem>
                    <SelectItem value="contacted">Contacted</SelectItem>
                    <SelectItem value="qualified">Qualified</SelectItem>
                    <SelectItem value="converted">Converted</SelectItem>
                    <SelectItem value="lost">Lost</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="notes">Notes</Label>
                <Textarea
                  id="notes"
                  value={editingLead.notes || ''}
                  onChange={(e) => setEditingLead({...editingLead, notes: e.target.value})}
                  rows={3}
                />
              </div>
              <Button onClick={handleSaveEdit}>Save Changes</Button>
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* View Dialog */}
      <Dialog open={!!viewingLead} onOpenChange={() => setViewingLead(null)}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>Lead Details</DialogTitle>
          </DialogHeader>
          {viewingLead && (
            <div className="grid gap-4 py-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-sm font-medium">Name</Label>
                  <p className="text-sm text-muted-foreground mt-1">
                    {viewingLead.name || 'Not provided'}
                  </p>
                </div>
                <div>
                  <Label className="text-sm font-medium">Status</Label>
                  <div className="mt-1">
                    <Badge 
                      variant={getStatusBadgeVariant(viewingLead.status)}
                      className="capitalize"
                    >
                      {viewingLead.status}
                    </Badge>
                  </div>
                </div>
                <div>
                  <Label className="text-sm font-medium">Phone</Label>
                  <p className="text-sm text-muted-foreground mt-1">
                    {viewingLead.phone || 'Not provided'}
                  </p>
                </div>
                <div>
                  <Label className="text-sm font-medium">Email</Label>
                  <p className="text-sm text-muted-foreground mt-1">
                    {viewingLead.email || 'Not provided'}
                  </p>
                </div>
                <div>
                  <Label className="text-sm font-medium">Source</Label>
                  <p className="text-sm text-muted-foreground mt-1 capitalize">
                    {viewingLead.source}
                  </p>
                </div>
                <div>
                  <Label className="text-sm font-medium">Campaign</Label>
                  <p className="text-sm text-muted-foreground mt-1">
                    {viewingLead.campaign_name || 'Not specified'}
                  </p>
                </div>
                <div>
                  <Label className="text-sm font-medium">Interest</Label>
                  <p className="text-sm text-muted-foreground mt-1">
                    {viewingLead.interest || 'Not specified'}
                  </p>
                </div>
                <div>
                  <Label className="text-sm font-medium">Created</Label>
                  <p className="text-sm text-muted-foreground mt-1">
                    {format(new Date(viewingLead.created_at), 'PPP')}
                  </p>
                </div>
              </div>
              {viewingLead.notes && (
                <div>
                  <Label className="text-sm font-medium">Notes</Label>
                  <p className="text-sm text-muted-foreground mt-1">
                    {viewingLead.notes}
                  </p>
                </div>
              )}
              {viewingLead.conversation_data && (
                <div>
                  <Label className="text-sm font-medium">Conversation Data</Label>
                  <pre className="text-xs bg-muted p-3 rounded mt-1 overflow-auto max-h-40">
                    {JSON.stringify(viewingLead.conversation_data, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
    </Card>
  )
}