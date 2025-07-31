import { useState } from 'react';
import { supabase } from '@/integrations/supabase/client';
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
      console.log('Calling MySQL API via Edge Function:', options);
      
      const { data, error } = await supabase.functions.invoke('mysql-api-bridge', {
        body: {
          endpoint: options.endpoint,
          method: options.method || 'GET',
          data: options.data,
          headers: options.headers
        }
      });

      if (error) {
        console.error('Edge Function error:', error);
        toast({
          title: "API Error",
          description: error.message || "Failed to call external API",
          variant: "destructive"
        });
        return { success: false, error: error.message };
      }

      if (!data.success) {
        console.error('External API error:', data.error);
        toast({
          title: "External API Error", 
          description: data.error || "External API returned an error",
          variant: "destructive"
        });
        return { success: false, error: data.error };
      }

      console.log('API call successful:', data);
      return {
        success: true,
        data: data.data,
        status: data.status
      };

    } catch (error: any) {
      console.error('Unexpected error:', error);
      toast({
        title: "Unexpected Error",
        description: "Something went wrong while calling the API",
        variant: "destructive"
      });
      return { success: false, error: error.message };
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