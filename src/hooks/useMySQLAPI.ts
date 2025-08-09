import { useState } from 'react';
import { useToast } from '@/hooks/use-toast';

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
      
      // Prepare the request payload
      const payload = {
        query: options.endpoint, // Using endpoint as SQL query
        params: options.data ? Object.values(options.data) : [],
        config: {
          host: '159.89.198.71',
          port: 3306,
          user: 'admin_aqil',
          password: 'admin_aqil',
          database: 'admin_railway'
        }
      };
      
      console.log('Sending payload:', JSON.stringify(payload));
      
      // Use direct API endpoint instead of Supabase Edge Function
      const response = await fetch('https://nodepath-chat-production.up.railway.app/mysql-api.php', {
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

      const result = await response.json();
      
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