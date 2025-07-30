import { useState, useEffect } from 'react'
import { supabase } from '@/integrations/supabase/client'
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
      let query = supabase.from('leads').select('*').order('created_at', { ascending: false })

      if (filters?.startDate) {
        query = query.gte('created_at', filters.startDate.toISOString())
      }
      if (filters?.endDate) {
        query = query.lte('created_at', filters.endDate.toISOString())
      }
      if (filters?.source) {
        query = query.eq('source', filters.source)
      }
      if (filters?.campaign) {
        query = query.eq('campaign_name', filters.campaign)
      }
      if (filters?.status) {
        query = query.eq('status', filters.status)
      }

      const { data, error } = await query
      if (error) throw error

      setLeads((data || []) as Lead[])
    } catch (error) {
      console.error('Error fetching leads:', error)
      toast.error('Failed to fetch leads')
    } finally {
      setLoading(false)
    }
  }

  const fetchStats = async (filters?: Partial<LeadFilters>) => {
    try {
      const { data, error } = await supabase.rpc('get_lead_stats', {
        start_date: filters?.startDate?.toISOString().split('T')[0] || new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        end_date: filters?.endDate?.toISOString().split('T')[0] || new Date().toISOString().split('T')[0],
        source_filter: filters?.source || null,
        campaign_filter: filters?.campaign || null
      })

      if (error) throw error
      setStats(data || [])
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
      const { data, error } = await supabase
        .from('leads')
        .insert([leadData])
        .select()
        .single()

      if (error) throw error
      
      toast.success('Lead created successfully')
      return data
    } catch (error) {
      console.error('Error creating lead:', error)
      toast.error('Failed to create lead')
      throw error
    }
  }

  const updateLead = async (id: string, updates: Partial<Lead>) => {
    try {
      const { data, error } = await supabase
        .from('leads')
        .update(updates)
        .eq('id', id)
        .select()
        .single()

      if (error) throw error
      
      toast.success('Lead updated successfully')
      return data
    } catch (error) {
      console.error('Error updating lead:', error)
      toast.error('Failed to update lead')
      throw error
    }
  }

  const deleteLead = async (id: string) => {
    try {
      const { error } = await supabase
        .from('leads')
        .delete()
        .eq('id', id)

      if (error) throw error
      
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