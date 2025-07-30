import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { LeadStats } from '@/types/leads'
import { format } from 'date-fns'

interface LeadChartProps {
  data: LeadStats[]
  loading: boolean
}

export const LeadChart = ({ data, loading }: LeadChartProps) => {
  // Transform data for chart display
  const chartData = data.reduce((acc, item) => {
    const date = format(new Date(item.period_date), 'MMM dd')
    const existing = acc.find(d => d.date === date)
    
    if (existing) {
      existing[item.source] = (existing[item.source] || 0) + item.total_leads
      existing.total = (existing.total || 0) + item.total_leads
    } else {
      acc.push({
        date,
        [item.source]: item.total_leads,
        total: item.total_leads
      })
    }
    
    return acc
  }, [] as any[])

  // Get unique sources for the legend
  const sources = [...new Set(data.map(item => item.source))]
  const sourceColors = {
    web: '#8884d8',
    whatsapp: '#25D366',
    instagram: '#E4405F',
    facebook: '#1877F2',
    default: '#82ca9d'
  }

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Lead Analytics</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-[400px] flex items-center justify-center">
            <div className="text-muted-foreground">Loading chart data...</div>
          </div>
        </CardContent>
      </Card>
    )
  }

  if (chartData.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Lead Analytics</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-[400px] flex items-center justify-center">
            <div className="text-muted-foreground">No data available</div>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Lead Analytics by Source</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-[400px]">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" />
              <YAxis />
              <Tooltip />
              <Legend />
              {sources.map(source => (
                <Bar 
                  key={source}
                  dataKey={source}
                  fill={sourceColors[source as keyof typeof sourceColors] || sourceColors.default}
                  name={source.charAt(0).toUpperCase() + source.slice(1)}
                />
              ))}
            </BarChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  )
}