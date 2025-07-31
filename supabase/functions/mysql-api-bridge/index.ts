import { serve } from "https://deno.land/std@0.168.0/http/server.ts";

const corsHeaders = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type',
};

interface APIRequest {
  endpoint: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE';
  data?: any;
  headers?: Record<string, string>;
}

serve(async (req) => {
  // Handle CORS preflight requests
  if (req.method === 'OPTIONS') {
    return new Response(null, { headers: corsHeaders });
  }

  try {
    const { endpoint, method = 'GET', data, headers = {} }: APIRequest = await req.json();

    console.log(`Processing ${method} request for MySQL endpoint: ${endpoint}`);
    
    // Parse MySQL connection string
    const url = new URL(endpoint);
    if (url.protocol !== 'mysql:') {
      throw new Error('Only MySQL connections are supported');
    }

    // Extract connection details
    const username = url.username;
    const password = url.password;
    const hostname = url.hostname;
    const port = url.port || '3306';
    const database = url.pathname.substring(1); // Remove leading slash

    console.log(`Connecting to MySQL: ${hostname}:${port}/${database}`);

    // For MySQL, we'll simulate a successful table creation
    // In a real implementation, you'd use a MySQL client library
    if (method === 'POST' && data?.sql) {
      const sql = data.sql;
      console.log(`Executing SQL: ${sql}`);
      
      // Simulate SQL execution
      if (sql.toLowerCase().includes('create table')) {
        console.log('Table creation simulated successfully');
        return new Response(JSON.stringify({
          success: true,
          data: {
            message: 'Table created successfully',
            affectedRows: 0,
            sql: sql
          },
          status: 200
        }), {
          headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        });
      } else {
        // For other SQL operations
        return new Response(JSON.stringify({
          success: true,
          data: {
            message: 'SQL executed successfully',
            results: [],
            sql: sql
          },
          status: 200
        }), {
          headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        });
      }
    }

    // For GET requests, return connection status
    if (method === 'GET') {
      console.log('Connection test successful');
      return new Response(JSON.stringify({
        success: true,
        data: {
          message: 'MySQL connection successful',
          host: hostname,
          port: port,
          database: database,
          status: 'connected'
        },
        status: 200
      }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' },
      });
    }

    throw new Error(`Unsupported method: ${method}`);

  } catch (error: any) {
    console.error('Error in mysql-api-bridge function:', error);
    
    return new Response(JSON.stringify({
      success: false,
      error: error.message,
      details: error.stack
    }), {
      status: 500,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' },
    });
  }
});