import { serve } from "https://deno.land/std@0.168.0/http/server.ts";
import { Client } from "https://deno.land/x/mysql@v2.12.1/mod.ts";

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
    const port = parseInt(url.port || '3306');
    const database = url.pathname.substring(1); // Remove leading slash

    console.log(`Connecting to MySQL: ${hostname}:${port}/${database}`);

    // Create MySQL client
    const client = await new Client().connect({
      hostname,
      username,
      password,
      port,
      db: database,
    });

    console.log('MySQL connection established');

    // Handle different operations
    if (method === 'POST' && data?.operation) {
      return await handleDatabaseOperation(client, data);
    }

    // Handle chunked data operations for large payloads
    if (method === 'POST' && data?.chunked_operation) {
      return await handleChunkedDatabaseOperation(client, data);
    }

    // Handle raw SQL execution
    if (method === 'POST' && data?.sql) {
      const sql = data.sql;
      console.log(`Executing SQL: ${sql}`);
      
      try {
        const result = await client.execute(sql);
        console.log('SQL executed successfully:', result);
        
        await client.close();
        
        return new Response(JSON.stringify({
          success: true,
          data: {
            message: 'SQL executed successfully',
            result: result,
            sql: sql
          },
          status: 200
        }), {
          headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        });
      } catch (error: any) {
        console.error('SQL execution error:', error);
        await client.close();
        
        return new Response(JSON.stringify({
          success: false,
          error: error.message,
          sql: sql
        }), {
          status: 500,
          headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        });
      }
    }

    // For GET requests, return connection status
    if (method === 'GET') {
      console.log('Connection test successful');
      await client.close();
      
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

async function handleDatabaseOperation(client: Client, data: any) {
  const { operation, table, payload, filters } = data;
  
  console.log(`Handling ${operation} operation on table ${table}`);
  
  // Check if table exists and create it if it doesn't
  await ensureTableExists(client, table);

  try {
    switch (operation) {
      case 'insert':
        console.log(`Inserting into ${table}:`, payload);
        
        // Generate ID if not provided
        if (!payload.id) {
          payload.id = generateId();
        }
        
        // Check if record with this ID already exists
        const checkSQL = `SELECT id FROM ${table} WHERE id = '${payload.id}'`;
        console.log('Check SQL:', checkSQL);
        const checkResult = await client.execute(checkSQL);
        
        if (checkResult.rows && checkResult.rows.length > 0) {
          console.log('Record exists, performing update instead');
          // Record exists, perform update
          const updateFields = Object.entries(payload)
            .filter(([key]) => key !== 'id')
            .map(([key, value]) => {
              if (typeof value === 'string') {
                // Convert ISO datetime to MySQL format
                if (value.match(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/)) {
                  const date = new Date(value);
                  const mysqlDateTime = date.toISOString().slice(0, 19).replace('T', ' ');
                  return `${key} = '${mysqlDateTime}'`;
                }
                // Properly escape JSON strings and quotes
                const escapedValue = value.replace(/\\/g, '\\\\').replace(/'/g, "\\'").replace(/"/g, '\\"');
                return `${key} = '${escapedValue}'`;
              } else if (typeof value === 'object') {
                // Handle JSON objects by stringifying and escaping
                const jsonString = JSON.stringify(value).replace(/\\/g, '\\\\').replace(/'/g, "\\'").replace(/"/g, '\\"');
                return `${key} = '${jsonString}'`;
              }
              return `${key} = ${value}`;
            })
            .join(', ');
          
          const updateSQL = `UPDATE ${table} SET ${updateFields} WHERE id = '${payload.id}'`;
          console.log('Update SQL:', updateSQL);
          const updateResult = await client.execute(updateSQL);
          
          await client.close();
          
          return new Response(JSON.stringify({
            success: true,
            data: {
              ...payload,
              affectedRows: updateResult.affectedRows
            }
          }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' },
          });
        } else {
          // Record doesn't exist, perform insert
          const columns = Object.keys(payload).join(', ');
          const values = Object.values(payload).map(v => {
            if (typeof v === 'string') {
              // Convert ISO datetime to MySQL format
              if (v.match(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/)) {
                const date = new Date(v);
                const mysqlDateTime = date.toISOString().slice(0, 19).replace('T', ' ');
                return `'${mysqlDateTime}'`;
              }
              // Properly escape JSON strings and quotes
              const escapedValue = v.replace(/\\/g, '\\\\').replace(/'/g, "\\'").replace(/"/g, '\\"');
              return `'${escapedValue}'`;
            } else if (typeof v === 'object') {
              // Handle JSON objects by stringifying and escaping
              const jsonString = JSON.stringify(v).replace(/\\/g, '\\\\').replace(/'/g, "\\'").replace(/"/g, '\\"');
              return `'${jsonString}'`;
            }
            return v;
          }).join(', ');
          const insertSQL = `INSERT INTO ${table} (${columns}) VALUES (${values})`;
          
          console.log('Insert SQL:', insertSQL);
          const insertResult = await client.execute(insertSQL);
          
          await client.close();
          
          return new Response(JSON.stringify({
            success: true,
            data: {
              ...payload,
              insertId: insertResult.lastInsertId,
              affectedRows: insertResult.affectedRows
            }
          }), {
            headers: { ...corsHeaders, 'Content-Type': 'application/json' },
          });
        }

      case 'select':
        console.log(`Selecting from ${table} with filters:`, filters);
        
        let selectSQL = `SELECT * FROM ${table}`;
        if (filters) {
          const conditions = Object.entries(filters)
            .filter(([key]) => key !== 'order_by')
            .map(([key, value]) => `${key} = '${value}'`)
            .join(' AND ');
          
          if (conditions) {
            selectSQL += ` WHERE ${conditions}`;
          }
          
          if (filters.order_by) {
            selectSQL += ` ORDER BY ${filters.order_by}`;
          }
        }
        
        console.log('Select SQL:', selectSQL);
        const selectResult = await client.execute(selectSQL);
        
        await client.close();
        
        return new Response(JSON.stringify({
          success: true,
          data: selectResult.rows || []
        }), {
          headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        });

      case 'update':
        console.log(`Updating ${table} with:`, payload);
        
        const updateFields = Object.entries(payload)
          .map(([key, value]) => `${key} = '${value}'`)
          .join(', ');
        
        let updateSQL = `UPDATE ${table} SET ${updateFields}`;
        if (filters) {
          const conditions = Object.entries(filters)
            .map(([key, value]) => `${key} = '${value}'`)
            .join(' AND ');
          updateSQL += ` WHERE ${conditions}`;
        }
        
        console.log('Update SQL:', updateSQL);
        const updateResult = await client.execute(updateSQL);
        
        await client.close();
        
        return new Response(JSON.stringify({
          success: true,
          data: {
            affectedRows: updateResult.affectedRows
          }
        }), {
          headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        });

      case 'delete':
        console.log(`Deleting from ${table} with filters:`, filters);
        
        let deleteSQL = `DELETE FROM ${table}`;
        if (filters) {
          const conditions = Object.entries(filters)
            .map(([key, value]) => `${key} = '${value}'`)
            .join(' AND ');
          deleteSQL += ` WHERE ${conditions}`;
        }
        
        console.log('Delete SQL:', deleteSQL);
        const deleteResult = await client.execute(deleteSQL);
        
        await client.close();
        
        return new Response(JSON.stringify({
          success: true,
          data: {
            affectedRows: deleteResult.affectedRows
          }
        }), {
          headers: { ...corsHeaders, 'Content-Type': 'application/json' },
        });

      default:
        await client.close();
        throw new Error(`Unsupported operation: ${operation}`);
    }
  } catch (error: any) {
    console.error(`Database operation error:`, error);
    await client.close();
    
    return new Response(JSON.stringify({
      success: false,
      error: error.message
    }), {
      status: 500,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' },
    });
  }
}

async function handleChunkedDatabaseOperation(client: Client, data: any) {
  const { operation, table, payload, chunk_size = 1000000 } = data; // 1MB chunks by default
  
  console.log(`Handling chunked ${operation} operation on table ${table}`);
  
  // Check if table exists and create it if it doesn't
  await ensureTableExists(client, table);

  try {
    if (operation === 'insert' || operation === 'update') {
      // Generate ID if not provided
      if (!payload.id) {
        payload.id = generateId();
      }

      // Process large fields in chunks
      const processedPayload = { ...payload };
      
      for (const [key, value] of Object.entries(payload)) {
        if (typeof value === 'string' && value.length > chunk_size) {
          console.log(`Processing large field ${key} with ${value.length} characters`);
          
          // Store large data in separate table with chunks
          const chunkTableName = `${table}_chunks`;
          await ensureChunkTableExists(client, chunkTableName);
          
          // Clear existing chunks for this record and field
          await client.execute(`DELETE FROM ${chunkTableName} WHERE record_id = '${payload.id}' AND field_name = '${key}'`);
          
          // Split data into chunks
          const chunks = [];
          for (let i = 0; i < value.length; i += chunk_size) {
            chunks.push(value.substring(i, i + chunk_size));
          }
          
          // Insert chunks
          for (let i = 0; i < chunks.length; i++) {
            const chunkId = `${payload.id}_${key}_${i}`;
            const chunkSQL = `INSERT INTO ${chunkTableName} (id, record_id, field_name, chunk_index, chunk_data) VALUES ('${chunkId}', '${payload.id}', '${key}', ${i}, '${chunks[i].replace(/'/g, "\\'")}')`;
            await client.execute(chunkSQL);
          }
          
          // Replace the large field with a reference
          processedPayload[key] = `CHUNKED:${chunks.length}`;
        }
      }

      // Proceed with normal operation using processed payload
      const modifiedData = { ...data, payload: processedPayload };
      return await handleDatabaseOperation(client, modifiedData);
    }

    throw new Error(`Chunked operation ${operation} not supported`);

  } catch (error: any) {
    console.error(`Chunked database operation error:`, error);
    await client.close();
    
    return new Response(JSON.stringify({
      success: false,
      error: error.message
    }), {
      status: 500,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' },
    });
  }
}

async function ensureChunkTableExists(client: Client, tableName: string) {
  console.log(`Checking if chunk table ${tableName} exists`);
  
  try {
    const checkResult = await client.execute(`SHOW TABLES LIKE '${tableName}'`);
    
    if (checkResult.rows && checkResult.rows.length > 0) {
      console.log(`Chunk table ${tableName} already exists`);
      return;
    }
    
    console.log(`Creating chunk table ${tableName}...`);
    
    const createChunkTableSQL = `
      CREATE TABLE IF NOT EXISTS ${tableName} (
        id VARCHAR(255) PRIMARY KEY,
        record_id VARCHAR(255) NOT NULL,
        field_name VARCHAR(255) NOT NULL,
        chunk_index INT NOT NULL,
        chunk_data LONGTEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        INDEX idx_record_field (record_id, field_name)
      )
    `;
    
    await client.execute(createChunkTableSQL);
    console.log(`Successfully created chunk table ${tableName}`);
  } catch (error: any) {
    console.error(`Error checking/creating chunk table ${tableName}:`, error);
  }
}

function generateId(): string {
  return `${Date.now()}-${Math.random().toString(36).substring(2)}`;
}

async function ensureTableExists(client: Client, tableName: string) {
  console.log(`Checking if table ${tableName} exists`);
  
  try {
    // Check if table exists
    const checkResult = await client.execute(`SHOW TABLES LIKE '${tableName}'`);
    
    if (checkResult.rows && checkResult.rows.length > 0) {
      console.log(`Table ${tableName} already exists`);
      return;
    }
    
    console.log(`Table ${tableName} does not exist, creating it...`);
    
    // Create table based on table name
    const createTableSQL = getCreateTableSQL(tableName);
    if (createTableSQL) {
      await client.execute(createTableSQL);
      console.log(`Successfully created table ${tableName}`);
    } else {
      console.log(`No table schema defined for ${tableName}`);
    }
  } catch (error: any) {
    console.error(`Error checking/creating table ${tableName}:`, error);
    // Don't throw here, let the main operation handle the error
  }
}

function getCreateTableSQL(tableName: string): string | null {
  const tableSchemas: Record<string, string> = {
    // Original tables with _nodepath suffix
    'chatbot_flows_nodepath': `
      CREATE TABLE IF NOT EXISTS chatbot_flows_nodepath (
        id VARCHAR(255) PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        description TEXT,
        nodes JSON NOT NULL,
        edges JSON NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
      )
    `,
    'chatbot_executions_nodepath': `
      CREATE TABLE IF NOT EXISTS chatbot_executions_nodepath (
        id VARCHAR(255) PRIMARY KEY,
        flow_id VARCHAR(255) NOT NULL,
        current_node_id VARCHAR(255) NOT NULL,
        variables JSON NOT NULL,
        messages JSON NOT NULL,
        is_waiting_for_input BOOLEAN NOT NULL DEFAULT FALSE,
        is_completed BOOLEAN NOT NULL DEFAULT FALSE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
      )
    `,
    'leads_nodepath': `
      CREATE TABLE IF NOT EXISTS leads_nodepath (
        id VARCHAR(255) PRIMARY KEY,
        name VARCHAR(255),
        email VARCHAR(255),
        phone VARCHAR(255),
        interest VARCHAR(255),
        status VARCHAR(255) DEFAULT 'new',
        source VARCHAR(255) NOT NULL DEFAULT 'web',
        campaign_name VARCHAR(255),
        flow_id VARCHAR(255),
        conversation_data JSON,
        notes TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
      )
    `,
    'media_files_nodepath': `
      CREATE TABLE IF NOT EXISTS media_files_nodepath (
        id VARCHAR(255) PRIMARY KEY,
        filename VARCHAR(255) NOT NULL,
        original_name VARCHAR(255) NOT NULL,
        file_type VARCHAR(255) NOT NULL,
        file_size BIGINT NOT NULL,
        storage_path VARCHAR(255) NOT NULL,
        public_url VARCHAR(255) NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
      )
    `,
    // Exact Supabase table replicas
    'chatbot_flows': `
      CREATE TABLE IF NOT EXISTS chatbot_flows (
        id VARCHAR(255) PRIMARY KEY,
        name TEXT NOT NULL,
        description LONGTEXT DEFAULT '',
        nodes LONGTEXT NOT NULL DEFAULT ('[]'),
        edges LONGTEXT NOT NULL DEFAULT ('[]'),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
      )
    `,
    'chatbot_executions': `
      CREATE TABLE IF NOT EXISTS chatbot_executions (
        id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
        flow_id TEXT NOT NULL,
        current_node_id TEXT NOT NULL,
        variables JSON NOT NULL DEFAULT ('{}'),
        messages JSON NOT NULL DEFAULT ('[]'),
        is_waiting_for_input BOOLEAN NOT NULL DEFAULT FALSE,
        is_completed BOOLEAN NOT NULL DEFAULT FALSE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
      )
    `,
    'leads': `
      CREATE TABLE IF NOT EXISTS leads (
        id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
        name TEXT,
        phone TEXT,
        email TEXT,
        interest TEXT,
        source TEXT NOT NULL DEFAULT 'web',
        campaign_name TEXT,
        flow_id TEXT,
        conversation_data JSON,
        status TEXT DEFAULT 'new',
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        notes TEXT
      )
    `,
    'media_files': `
      CREATE TABLE IF NOT EXISTS media_files (
        id VARCHAR(36) PRIMARY KEY DEFAULT (UUID()),
        filename TEXT NOT NULL,
        original_name TEXT NOT NULL,
        file_type TEXT NOT NULL,
        file_size BIGINT NOT NULL,
        storage_path TEXT NOT NULL,
        public_url TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      )
    `
  };
  
  return tableSchemas[tableName] || null;
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