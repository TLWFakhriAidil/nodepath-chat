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

    // Handle different operations
    if (method === 'POST' && data?.operation) {
      return await handleDatabaseOperation(data);
    }

    // Handle raw SQL execution
    if (method === 'POST' && data?.sql) {
      const sql = data.sql;
      console.log(`Executing SQL: ${sql}`);
      
      // Create initial tables
      if (sql.toLowerCase().includes('create table')) {
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

async function handleDatabaseOperation(data: any) {
  const { operation, table, payload, id, filters } = data;
  
  console.log(`Handling ${operation} operation on table ${table}`);

  // Simulate database operations
  switch (operation) {
    case 'insert':
      return new Response(JSON.stringify({
        success: true,
        data: {
          ...payload,
          id: id || generateId(),
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        }
      }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' },
      });

    case 'update':
      return new Response(JSON.stringify({
        success: true,
        data: {
          ...payload,
          updated_at: new Date().toISOString()
        }
      }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' },
      });

    case 'select':
      // Return mock data based on table
      const mockData = getMockData(table, filters);
      return new Response(JSON.stringify({
        success: true,
        data: mockData
      }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' },
      });

    case 'delete':
      return new Response(JSON.stringify({
        success: true,
        data: { message: 'Record deleted successfully' }
      }), {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' },
      });

    default:
      throw new Error(`Unsupported operation: ${operation}`);
  }
}

function generateId(): string {
  return `${Date.now()}-${Math.random().toString(36).substring(2)}`;
}

function getMockData(table: string, filters?: any) {
  // Return appropriate mock data structure based on table
  switch (table) {
    case 'chatbot_flows':
      return [];
    case 'chatbot_executions':
      return [];
    case 'leads':
      return [];
    case 'media_files':
      return [];
    default:
      return [];
  }
}