-- Create leads table for tracking chatbot conversations
CREATE TABLE IF NOT EXISTS public.leads (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  name TEXT,
  phone TEXT,
  email TEXT,
  interest TEXT,
  source TEXT NOT NULL DEFAULT 'web', -- e.g., 'whatsapp', 'web', 'instagram'
  campaign_name TEXT,
  flow_id TEXT, -- Reference to the chatbot flow that generated this lead
  conversation_data JSONB, -- Store the full conversation context
  status TEXT DEFAULT 'new' CHECK (status IN ('new', 'contacted', 'qualified', 'converted', 'lost')),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  notes TEXT
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_leads_created_at ON public.leads(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_leads_source ON public.leads(source);
CREATE INDEX IF NOT EXISTS idx_leads_campaign_name ON public.leads(campaign_name);
CREATE INDEX IF NOT EXISTS idx_leads_status ON public.leads(status);
CREATE INDEX IF NOT EXISTS idx_leads_flow_id ON public.leads(flow_id);

-- Enable Row Level Security
ALTER TABLE public.leads ENABLE ROW LEVEL SECURITY;

-- Create policies for leads table (allowing all operations for now)
CREATE POLICY "Allow all operations for everyone" ON public.leads
  FOR ALL USING (true);

-- Create trigger for automatic timestamp updates
CREATE TRIGGER update_leads_updated_at
    BEFORE UPDATE ON public.leads
    FOR EACH ROW
    EXECUTE FUNCTION public.update_updated_at_column();

-- Create a function to get lead statistics
CREATE OR REPLACE FUNCTION public.get_lead_stats(
  start_date DATE DEFAULT CURRENT_DATE - INTERVAL '30 days',
  end_date DATE DEFAULT CURRENT_DATE,
  source_filter TEXT DEFAULT NULL,
  campaign_filter TEXT DEFAULT NULL
)
RETURNS TABLE(
  period_date DATE,
  total_leads BIGINT,
  source TEXT,
  campaign_name TEXT
) AS $$
BEGIN
  RETURN QUERY
  SELECT 
    DATE(l.created_at) as period_date,
    COUNT(*) as total_leads,
    l.source,
    l.campaign_name
  FROM public.leads l
  WHERE 
    DATE(l.created_at) >= start_date 
    AND DATE(l.created_at) <= end_date
    AND (source_filter IS NULL OR l.source = source_filter)
    AND (campaign_filter IS NULL OR l.campaign_name = campaign_filter)
  GROUP BY DATE(l.created_at), l.source, l.campaign_name
  ORDER BY period_date DESC, l.source, l.campaign_name;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;