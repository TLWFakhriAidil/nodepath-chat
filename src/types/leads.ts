export interface Lead {
  id: string
  name?: string
  phone?: string
  email?: string
  interest?: string
  source: string
  campaign_name?: string
  flow_id?: string
  conversation_data?: any
  status: 'new' | 'contacted' | 'qualified' | 'converted' | 'lost'
  created_at: string
  updated_at: string
  notes?: string
}

export interface LeadStats {
  period_date: string
  total_leads: number
  source: string
  campaign_name?: string
}

export interface LeadFilters {
  startDate: Date
  endDate: Date
  source?: string
  campaign?: string
  status?: string
}

export interface LeadSummary {
  totalLeads: number
  newLeads: number
  convertedLeads: number
  conversionRate: number
  topSource: string
  topCampaign?: string
}