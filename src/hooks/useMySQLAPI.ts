import { useState } from 'react';
import { useToast } from '@/hooks/use-toast';

// Function to fetch database configuration from backend
const fetchDatabaseConfig = async () => {
  try {
    const response = await fetch('/api/config/database');
    if (response.ok) {
      return await response.json();
    }
  } catch (error) {
    console.warn('Failed to fetch database config from backend, using fallback');
  }
  
  // Fallback configuration (should not contain real credentials)
  return {
    host: 'localhost',
    port: 3306,
    user: 'root',
    password: '',
    database: 'admin_railway'
  };
}

interface APICallOptions {
  endpoint: string;
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  data?: any;
  headers?: Record<string, string>;
}

interface APIResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  status?: number;
}

export const useMySQLAPI = () => {
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  const callAPI = async <T = any>(options: APICallOptions): Promise<APIResponse<T>> => {
    setLoading(true);
    
    try {
      console.log('MySQL API call via Supabase Edge Function:', options);
      
      // Fetch database configuration from backend
      const config = await fetchDatabaseConfig();
      
      // Prepare the request payload
      const payload = {
        query: options.endpoint, // Using endpoint as SQL query
        params: options.data ? Object.values(options.data) : [],
        config
      };
      
      console.log('Sending payload:', JSON.stringify(payload));
      
      // Use local PHP endpoint for MySQL operations
      const response = await fetch('/mysql-api.php', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          ...options.headers
        },
        body: JSON.stringify(payload)
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      // Get response as text first to handle potential JSON parsing issues
      const responseText = await response.text();
      
      // Check if response is empty
      if (!responseText) {
        throw new Error('Empty response from server');
      }
      
      // Try to parse the response as JSON
      let result;
      try {
        result = JSON.parse(responseText);
      } catch (parseError) {
        console.error('Failed to parse JSON response:', responseText);
        throw new Error(`JSON parse error: ${parseError.message}`);
      }
      
      if (result.success) {
        toast({
          title: "MySQL Success",
          description: "MySQL operation completed successfully",
        });
        return { success: true, data: result.data, status: response.status };
      } else {
        throw new Error(result.error || 'MySQL operation failed');
      }

    } catch (error: any) {
      console.error('MySQL API error:', error);
      toast({
        title: "MySQL Error",
        description: error.message || "Failed to connect to MySQL database",
        variant: "destructive"
      });
      return { success: false, error: error.message, status: 500 };
    } finally {
      setLoading(false);
    }
  };

  // Convenience methods for common operations
  const get = <T = any>(endpoint: string, headers?: Record<string, string>) => 
    callAPI<T>({ endpoint, method: 'GET', headers });

  const post = <T = any>(endpoint: string, data?: any, headers?: Record<string, string>) => 
    callAPI<T>({ endpoint, method: 'POST', data, headers });

  const put = <T = any>(endpoint: string, data?: any, headers?: Record<string, string>) => 
    callAPI<T>({ endpoint, method: 'PUT', data, headers });

  const del = <T = any>(endpoint: string, headers?: Record<string, string>) => 
    callAPI<T>({ endpoint, method: 'DELETE', headers });

  return {
    callAPI,
    get,
    post,
    put,
    delete: del,
    loading
  };
};