import { serve } from "https://deno.land/std@0.168.0/http/server.ts"

const corsHeaders = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type',
}

serve(async (req) => {
  // Handle CORS preflight requests
  if (req.method === 'OPTIONS') {
    return new Response('ok', { headers: corsHeaders })
  }

  try {
    const { query, params } = await req.json()

    // Your MySQL connection configuration
    const mysqlConfig = {
      hostname: '159.89.198.71',
      port: 3306,
      username: 'admin_aqil',
      password: 'admin_aqil',
      db: 'admin_railway',
    }

    // MySQL connection using Deno MySQL client
    const mysql = await import('https://deno.land/x/mysql@v2.12.1/mod.ts')
    
    const client = await new mysql.Client().connect(mysqlConfig)

    let result
    // Handle multiple SQL statements (separated by semicolon)
    const statements = query.split(';').filter(stmt => stmt.trim())
    
    if (statements.length > 1) {
      // Execute multiple statements
      for (let i = 0; i < statements.length; i++) {
        const stmt = statements[i].trim()
        if (stmt) {
          if (i === statements.length - 1 && params && params.length > 0) {
            // Last statement with parameters
            result = await client.execute(stmt, params)
          } else {
            // Other statements without parameters
            result = await client.execute(stmt)
          }
        }
      }
    } else {
      // Single statement
      if (params && params.length > 0) {
        result = await client.execute(query, params)
      } else {
        result = await client.execute(query)
      }
    }

    await client.close()

    console.log('MySQL operation successful:', { query, params, rowCount: result.affectedRows })

    return new Response(
      JSON.stringify({ 
        success: true, 
        data: result.rows || result,
        affectedRows: result.affectedRows || 0
      }),
      {
        headers: { ...corsHeaders, 'Content-Type': 'application/json' },
      },
    )
  } catch (error) {
    console.error('MySQL API Error:', error)
    return new Response(
      JSON.stringify({ 
        success: false, 
        error: error.message 
      }),
      {
        status: 500,
        headers: { ...corsHeaders, 'Content-Type': 'application/json' },
      },
    )
  }
})