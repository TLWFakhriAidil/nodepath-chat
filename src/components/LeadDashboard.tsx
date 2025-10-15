import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { DatePickerWithRange } from '@/components/ui/date-picker'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Download, Users, TrendingUp, Target, Filter } from 'lucide-react'
import { useLeads } from '@/hooks/useLeads'
import { LeadFilters } from '@/types/leads'
import { LeadChart } from '@/components/LeadChart'
import { LeadTable } from '@/components/LeadTable'
import { useDevice } from '@/contexts/DeviceContext'
import DeviceRequiredPopup from '@/components/DeviceRequiredPopup'
import { addDays } from 'date-fns'

/**
 * Lead Dashboard component with device-based filtering
 * Automatically filters leads by user's configured devices
 */
export const LeadDashboard = () => {
  const { has_devices } = useDevice()
  const { 
    leads, 
    stats, 
    summary, 
    loading, 
    deviceRequiredError,
    fetchLeads, 
    fetchStats, 
    generateSummary, 
    exportToCSV 
  } = useLeads()
  const [showDeviceRequiredPopup, setShowDeviceRequiredPopup] = useState(false)

  const [filters, setFilters] = useState<Partial<LeadFilters>>({
    startDate: addDays(new Date(), -30),
    endDate: new Date()
  })

  useEffect(() => {
    fetchLeads(filters)
    fetchStats(filters)
  }, [])

  useEffect(() => {
    generateSummary(leads)
  }, [leads])

  // Show device required popup when deviceRequiredError is true
  useEffect(() => {
    if (deviceRequiredError) {
      setShowDeviceRequiredPopup(true)
    }
  }, [deviceRequiredError])

  const handleFilterChange = (key: keyof LeadFilters, value: any) => {
    const newFilters = { ...filters, [key]: value }
    setFilters(newFilters)
  }

  const applyFilters = () => {
    fetchLeads(filters)
    fetchStats(filters)
  }

  const clearFilters = () => {
    const defaultFilters = {
      startDate: addDays(new Date(), -30),
      endDate: new Date()
    }
    setFilters(defaultFilters)
    fetchLeads(defaultFilters)
    fetchStats(defaultFilters)
  }

  const handleExport = () => {
    exportToCSV(leads)
  }

  return (
    <div className="p-4 md:p-6 lg:p-8 space-y-6 max-w-7xl mx-auto">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <h1 className="text-2xl md:text-3xl font-bold">Lead Analytics Dashboard</h1>
        <Button onClick={handleExport} className="flex items-center gap-2 w-fit">
          <Download className="h-4 w-4" />
          Export CSV
        </Button>
      </div>

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Filter className="h-5 w-5" />
            Filters
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="sm:col-span-2 lg:col-span-1">
              <label className="text-sm font-medium mb-2 block">Date Range</label>
              <DatePickerWithRange
                from={filters.startDate}
                to={filters.endDate}
                onSelect={(range) => {
                  if (range?.from) handleFilterChange('startDate', range.from)
                  if (range?.to) handleFilterChange('endDate', range.to)
                }}
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">Source</label>
              <Select
                value={filters.source || 'all'}
                onValueChange={(value) => handleFilterChange('source', value === 'all' ? undefined : value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="All Sources" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Sources</SelectItem>
                  <SelectItem value="web">Web</SelectItem>
                  <SelectItem value="whatsapp">WhatsApp</SelectItem>
                  <SelectItem value="instagram">Instagram</SelectItem>
                  <SelectItem value="facebook">Facebook</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <label className="text-sm font-medium mb-2 block">Status</label>
              <Select
                value={filters.status || 'all'}
                onValueChange={(value) => handleFilterChange('status', value === 'all' ? undefined : value)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="All Statuses" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Statuses</SelectItem>
                  <SelectItem value="new">New</SelectItem>
                  <SelectItem value="contacted">Contacted</SelectItem>
                  <SelectItem value="qualified">Qualified</SelectItem>
                  <SelectItem value="converted">Converted</SelectItem>
                  <SelectItem value="lost">Lost</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="sm:col-span-2 lg:col-span-1 flex items-end gap-2">
              <Button onClick={applyFilters} className="flex-1">
                Apply Filters
              </Button>
              <Button variant="outline" onClick={clearFilters}>
                Clear
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Summary Cards */}
      {summary && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Leads</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{summary.totalLeads}</div>
              <p className="text-xs text-muted-foreground">
                {summary.newLeads} new leads
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Conversion Rate</CardTitle>
              <TrendingUp className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{summary.conversionRate.toFixed(1)}%</div>
              <p className="text-xs text-muted-foreground">
                {summary.convertedLeads} converted
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Top Source</CardTitle>
              <Target className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{summary.topSource}</div>
              <p className="text-xs text-muted-foreground">
                Most leads from this source
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Top Campaign</CardTitle>
              <Target className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{summary.topCampaign || 'N/A'}</div>
              <p className="text-xs text-muted-foreground">
                Best performing campaign
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Charts and Table */}
      <Tabs defaultValue="chart" className="space-y-4">
        <TabsList>
          <TabsTrigger value="chart">Analytics Chart</TabsTrigger>
          <TabsTrigger value="table">Lead Details</TabsTrigger>
        </TabsList>

        <TabsContent value="chart" className="space-y-4">
          <LeadChart data={stats} loading={loading} />
        </TabsContent>

        <TabsContent value="table" className="space-y-4">
          <LeadTable data={leads} loading={loading} onRefresh={() => fetchLeads(filters)} />
        </TabsContent>
      </Tabs>
      
      {/* Device Required Popup */}
      <DeviceRequiredPopup 
        open={showDeviceRequiredPopup} 
        onOpenChange={setShowDeviceRequiredPopup} 
      />
    </div>
  )
}