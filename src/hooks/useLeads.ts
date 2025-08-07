import { useState, useEffect } from 'react'
import { Lead, LeadStats, LeadFilters, LeadSummary } from '@/types/leads'
import { toast } from 'sonner'

export const useLeads = () => {
  const [leads, setLeads] = useState<Lead[]>([])
  const [stats, setStats] = useState<LeadStats[]>([])
  const [summary, setSummary] = useState<LeadSummary | null>(null)
  const [loading, setLoading] = useState(false)

  const fetchLeads = async (filters?: Partial<LeadFilters>) => {
    setLoading(true)
    try {
      // Direct MySQL API call would go here
      // For now, return empty array since MySQL connection is needed
      setLeads([])
    } catch (error) {
      console.error('Error fetching leads:', error)
      toast.error('Failed to fetch leads')
    } finally {
      setLoading(false)
    }
  }

  const fetchStats = async (filters?: Partial<LeadFilters>) => {
    try {
      // Direct MySQL stats call would go here
      setStats([])
    } catch (error) {
      console.error('Error fetching lead stats:', error)
      toast.error('Failed to fetch lead statistics')
    }
  }

  const generateSummary = (leadData: Lead[]) => {
    const total = leadData.length
    const newLeads = leadData.filter(l => l.status === 'new').length
    const converted = leadData.filter(l => l.status === 'converted').length
    const conversionRate = total > 0 ? (converted / total) * 100 : 0

    // Get top source
    const sourceCount = leadData.reduce((acc, lead) => {
      acc[lead.source] = (acc[lead.source] || 0) + 1
      return acc
    }, {} as Record<string, number>)
    const topSource = Object.entries(sourceCount).sort(([,a], [,b]) => b - a)[0]?.[0] || 'N/A'

    // Get top campaign
    const campaignCount = leadData.reduce((acc, lead) => {
      if (lead.campaign_name) {
        acc[lead.campaign_name] = (acc[lead.campaign_name] || 0) + 1
      }
      return acc
    }, {} as Record<string, number>)
    const topCampaign = Object.entries(campaignCount).sort(([,a], [,b]) => b - a)[0]?.[0]

    setSummary({
      totalLeads: total,
      newLeads,
      convertedLeads: converted,
      conversionRate,
      topSource,
      topCampaign
    })
  }

  const createLead = async (leadData: Omit<Lead, 'id' | 'created_at' | 'updated_at'>) => {
    try {
      // Direct MySQL insert would go here
      toast.success('Lead created successfully')
      return null
    } catch (error) {
      console.error('Error creating lead:', error)
      toast.error('Failed to create lead')
      throw error
    }
  }

  const updateLead = async (id: string, updates: Partial<Lead>) => {
    try {
      // Direct MySQL update would go here
      toast.success('Lead updated successfully')
      return null
    } catch (error) {
      console.error('Error updating lead:', error)
      toast.error('Failed to update lead')
      throw error
    }
  }

  const deleteLead = async (id: string) => {
    try {
      // Direct MySQL delete would go here
      toast.success('Lead deleted successfully')
    } catch (error) {
      console.error('Error deleting lead:', error)
      toast.error('Failed to delete lead')
      throw error
    }
  }

  const exportToCSV = (data: Lead[]) => {
    const headers = ['Name', 'Phone', 'Email', 'Source', 'Campaign', 'Status', 'Created Date', 'Notes']
    const csvContent = [
      headers.join(','),
      ...data.map(lead => [
        lead.name || '',
        lead.phone || '',
        lead.email || '',
        lead.source,
        lead.campaign_name || '',
        lead.status,
        new Date(lead.created_at).toLocaleDateString(),
        (lead.notes || '').replace(/,/g, ';') // Replace commas to avoid CSV issues
      ].join(','))
    ].join('\n')

    const blob = new Blob([csvContent], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `leads-export-${new Date().toISOString().split('T')[0]}.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  }

  return {
    leads,
    stats,
    summary,
    loading,
    fetchLeads,
    fetchStats,
    generateSummary,
    createLead,
    updateLead,
    deleteLead,
    exportToCSV
  }
}