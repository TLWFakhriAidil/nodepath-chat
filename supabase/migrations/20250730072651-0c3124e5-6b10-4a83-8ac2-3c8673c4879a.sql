-- Fix the search path security issue for the get_lead_stats function
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
$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = 'public';