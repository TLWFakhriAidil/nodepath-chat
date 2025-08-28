import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { useMySQLAPI } from '@/hooks/useMySQLAPI';
import { Loader2, Database, Send, Settings } from 'lucide-react';
import { saveFlow, getFlows } from '@/lib/mysqlStorage'

export default function MySQLAPIExample() {
  const { get, post, loading } = useMySQLAPI();
  const [endpoint, setEndpoint] = useState('mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway');
  const [postData, setPostData] = useState('{"sql": "CREATE TABLE IF NOT EXISTS chatbot_flows (id VARCHAR(255) PRIMARY KEY, name VARCHAR(255) NOT NULL, description TEXT, nodes JSON NOT NULL DEFAULT \'[]\', edges JSON NOT NULL DEFAULT \'[]\', created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP)"}');
  const [response, setResponse] = useState<any>(null);

  // Example: Fetch data on component mount
  useEffect(() => {
    const fetchInitialData = async () => {
      const result = await get('mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway');
      if (result.success) {
        console.log('Initial data loaded:', result.data);
      }
    };
    
    // Uncomment to auto-fetch on load
    // fetchInitialData();
  }, []);

  const handleGetRequest = async () => {
    const result = await get(endpoint);
    setResponse(result);
  };

  const handlePostRequest = async () => {
    try {
      const data = JSON.parse(postData);
      const result = await post(endpoint, data);
      setResponse(result);
    } catch (error) {
      setResponse({ success: false, error: 'Invalid JSON in request body' });
    }
  };

  const testConnection = async () => {
    try {
      // Test direct connection to the MySQL API
      const response = await fetch('https://nodepath-chat-production.up.railway.app/mysql-api.php', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
        body: JSON.stringify({
          query: 'SELECT 1 as test',
          params: [],
          config: {
            host: '157.245.206.124',
            port: 3306,
            user: 'admin_aqil',
            password: 'admin_aqil',
            database: 'admin_railway'
          }
        })
      });
      
      // Get response as text first to handle potential JSON parsing issues
      const responseText = await response.text();
      console.log('Raw API response:', responseText);
      
      // Check if response is empty
      if (!responseText) {
        setResponse({ 
          success: false, 
          error: 'Empty response from server'
        });
        return;
      }
      
      // Try to parse the response as JSON
      try {
        const jsonResult = JSON.parse(responseText);
        setResponse(jsonResult);
      } catch (parseError) {
        console.error('Failed to parse JSON response:', responseText);
        setResponse({ 
          success: false, 
          error: 'Failed to parse response: ' + parseError.message,
          rawResponse: responseText
        });
      }
    } catch (error: any) {
      setResponse({ success: false, error: error.message })
    }
  }

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Database className="w-5 h-5" />
            MySQL API Bridge Example
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <Label htmlFor="endpoint">API Endpoint</Label>
            <Input
              id="endpoint"
              value={endpoint}
              onChange={(e) => setEndpoint(e.target.value)}
              placeholder="mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway"
              className="mt-1"
            />
          </div>

          <div className="flex gap-2">
            <Button 
              onClick={handleGetRequest} 
              disabled={loading}
              variant="default"
            >
              {loading ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <Send className="w-4 h-4 mr-2" />}
              GET Request
            </Button>
            <Button 
              onClick={testConnection} 
              disabled={loading}
              variant="outline"
            >
              {loading ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <Settings className="w-4 h-4 mr-2" />}
              Test Connection
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>POST Request</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <Label htmlFor="postData">Request Body (JSON)</Label>
            <Textarea
              id="postData"
              value={postData}
              onChange={(e) => setPostData(e.target.value)}
              placeholder='{"key": "value"}'
              className="mt-1 font-mono"
              rows={4}
            />
          </div>

          <Button 
            onClick={handlePostRequest} 
            disabled={loading}
            variant="secondary"
          >
            {loading ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <Send className="w-4 h-4 mr-2" />}
            POST Request
          </Button>
        </CardContent>
      </Card>

      {response && (
        <Card>
          <CardHeader>
            <CardTitle>API Response</CardTitle>
          </CardHeader>
          <CardContent>
            <pre className="bg-muted p-4 rounded-lg overflow-auto text-sm">
              {JSON.stringify(response, null, 2)}
            </pre>
          </CardContent>
        </Card>
      )}
    </div>
  );
}