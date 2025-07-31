import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { useMySQLAPI } from '@/hooks/useMySQLAPI';
import { Loader2, Database, Send } from 'lucide-react';

export default function MySQLAPIExample() {
  const { get, post, loading } = useMySQLAPI();
  const [endpoint, setEndpoint] = useState('https://myapi.example.com/users');
  const [postData, setPostData] = useState('{"name": "John Doe", "email": "john@example.com"}');
  const [response, setResponse] = useState<any>(null);

  // Example: Fetch data on component mount
  useEffect(() => {
    const fetchInitialData = async () => {
      const result = await get('https://myapi.example.com/users');
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
              placeholder="https://myapi.example.com/users"
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