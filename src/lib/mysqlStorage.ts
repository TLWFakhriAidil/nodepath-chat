import { ChatbotFlow, MediaFile, FlowExecution } from '@/types/chatbot'
import { supabase } from '@/integrations/supabase/client'

const MYSQL_ENDPOINT = 'mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway'

// MySQL API helper function
const callMySQLAPI = async (data: any) => {
  const { data: result, error } = await supabase.functions.invoke('mysql-api-bridge', {
    body: {
      endpoint: MYSQL_ENDPOINT,
      method: 'POST',
      data
    }
  })
  
  if (error) throw error
  return result
}

// Flow management
export const saveFlow = async (flow: ChatbotFlow): Promise<void> => {
  try {
    const result = await callMySQLAPI({
      operation: 'insert',
      table: 'chatbot_flows_nodepath',
      payload: {
        id: flow.id,
        name: flow.name,
        description: flow.description,
        nodes: JSON.stringify(flow.nodes),
        edges: JSON.stringify(flow.edges),
        updated_at: new Date().toISOString()
      }
    })
    
    if (!result.success) throw new Error(result.error)
  } catch (error) {
    console.error('Error saving flow to MySQL:', error)
    throw error
  }
}

export const getFlows = async (): Promise<ChatbotFlow[]> => {
  try {
    const result = await callMySQLAPI({
      operation: 'select',
      table: 'chatbot_flows_nodepath',
      filters: { order_by: 'updated_at DESC' }
    })
    
    if (!result.success) throw new Error(result.error)
    
    return (result.data || []).map((row: any) => ({
      id: row.id,
      name: row.name,
      description: row.description,
      nodes: JSON.parse(row.nodes || '[]'),
      edges: JSON.parse(row.edges || '[]'),
      createdAt: row.created_at,
      updatedAt: row.updated_at
    }))
  } catch (error) {
    console.error('Error fetching flows from MySQL:', error)
    return []
  }
}

export const getFlow = async (id: string): Promise<ChatbotFlow | null> => {
  try {
    const result = await callMySQLAPI({
      operation: 'select',
      table: 'chatbot_flows_nodepath',
      filters: { id }
    })
    
    if (!result.success) throw new Error(result.error)
    if (!result.data || result.data.length === 0) return null
    
    const row = result.data[0]
    return {
      id: row.id,
      name: row.name,
      description: row.description,
      nodes: JSON.parse(row.nodes || '[]'),
      edges: JSON.parse(row.edges || '[]'),
      createdAt: row.created_at,
      updatedAt: row.updated_at
    }
  } catch (error) {
    console.error('Error fetching flow from MySQL:', error)
    return null
  }
}

export const deleteFlow = async (id: string): Promise<void> => {
  try {
    const result = await callMySQLAPI({
      operation: 'delete',
      table: 'chatbot_flows_nodepath',
      filters: { id }
    })
    
    if (!result.success) throw new Error(result.error)
  } catch (error) {
    console.error('Error deleting flow from MySQL:', error)
    throw error
  }
}

// Helper function to clean and escape JSON strings for MySQL
const cleanJsonForMySQL = (obj: any): string => {
  // First stringify the object
  let jsonString = JSON.stringify(obj);
  
  // Clean up any problematic characters that might cause encoding issues
  jsonString = jsonString
    .replace(/\\t/g, ' ')      // Replace tab characters with spaces
    .replace(/\\n/g, ' ')      // Replace newlines with spaces
    .replace(/\\r/g, ' ')      // Replace carriage returns with spaces
    .replace(/\t/g, ' ')       // Replace actual tab characters
    .replace(/\n/g, ' ')       // Replace actual newlines
    .replace(/\r/g, ' ')       // Replace actual carriage returns
    .replace(/[\x00-\x1F\x7F]/g, ' '); // Replace control characters with spaces
  
  return jsonString;
}

// Execution management
export const saveExecution = async (execution: FlowExecution): Promise<void> => {
  try {
    // Clean the messages and variables for MySQL storage
    const cleanedMessages = execution.messages.map(msg => ({
      ...msg,
      content: msg.content?.replace(/[\x00-\x1F\x7F]/g, ' ') || '',
      mediaUrl: msg.mediaUrl?.replace(/[\x00-\x1F\x7F]/g, ' ') || undefined
    }));

    const executionData = {
      flow_id: execution.flowId,
      current_node_id: execution.currentNodeId,
      variables: cleanJsonForMySQL(execution.variables),
      messages: cleanJsonForMySQL(cleanedMessages),
      is_waiting_for_input: execution.isWaitingForInput,
      is_completed: execution.isCompleted
    }

    console.log('Saving/updating execution to MySQL for flow:', execution.flowId)
    console.log('Messages to save:', cleanedMessages.length, 'messages')
    
    // First check if execution exists for this flow_id
    const existingResult = await callMySQLAPI({
      operation: 'select',
      table: 'chatbot_executions_nodepath',
      filters: { flow_id: execution.flowId }
    })

    if (existingResult.success && existingResult.data && existingResult.data.length > 0) {
      // Update existing record
      const result = await callMySQLAPI({
        operation: 'update',
        table: 'chatbot_executions_nodepath',
        filters: { flow_id: execution.flowId },
        payload: executionData
      })

      if (!result.success) throw new Error(result.error)
      console.log('Execution updated successfully for flow:', execution.flowId)
    } else {
      // Insert new record with flow_id as unique identifier
      const result = await callMySQLAPI({
        operation: 'insert',
        table: 'chatbot_executions_nodepath',
        payload: {
          id: execution.flowId, // Use flow_id as the primary key
          ...executionData
        }
      })

      if (!result.success) throw new Error(result.error)
      console.log('Execution saved successfully for flow:', execution.flowId)
    }
  } catch (error) {
    console.error('Error saving execution to MySQL:', error)
    throw error
  }
}

export const getExecution = async (flowId: string): Promise<FlowExecution | null> => {
  try {
    const result = await callMySQLAPI({
      operation: 'select',
      table: 'chatbot_executions_nodepath',
      filters: { flow_id: flowId }
    })
    
    if (!result.success) throw new Error(result.error)
    if (!result.data || result.data.length === 0) return null
    
    const row = result.data[0]
    return {
      flowId: row.flow_id,
      currentNodeId: row.current_node_id,
      variables: JSON.parse(row.variables || '{}'),
      messages: JSON.parse(row.messages || '[]'),
      isWaitingForInput: row.is_waiting_for_input,
      isCompleted: row.is_completed
    }
  } catch (error) {
    console.error('Error fetching execution from MySQL:', error)
    return null
  }
}

export const getExecutions = async (): Promise<FlowExecution[]> => {
  try {
    const result = await callMySQLAPI({
      operation: 'select',
      table: 'chatbot_executions_nodepath',
      filters: { order_by: 'updated_at DESC' }
    })
    
    if (!result.success) throw new Error(result.error)
    
    return (result.data || []).map((row: any) => ({
      flowId: row.flow_id,
      currentNodeId: row.current_node_id,
      variables: JSON.parse(row.variables || '{}'),
      messages: JSON.parse(row.messages || '[]'),
      isWaitingForInput: row.is_waiting_for_input,
      isCompleted: row.is_completed
    }))
  } catch (error) {
    console.error('Error fetching executions from MySQL:', error)
    return []
  }
}

export const deleteExecution = async (flowId: string): Promise<void> => {
  try {
    const result = await callMySQLAPI({
      operation: 'delete',
      table: 'chatbot_executions_nodepath',
      filters: { flow_id: flowId }
    })
    
    if (!result.success) throw new Error(result.error)
  } catch (error) {
    console.error('Error deleting execution from MySQL:', error)
    throw error
  }
}

// Initialize MySQL tables
export const initializeMySQLTables = async (): Promise<void> => {
  const tables = [
    // Chatbot flows table
    `CREATE TABLE IF NOT EXISTS chatbot_flows_nodepath (
      id VARCHAR(255) PRIMARY KEY,
      name VARCHAR(255) NOT NULL,
      description TEXT,
      nodes JSON NOT NULL,
      edges JSON NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    )`,
    
    // Chatbot executions table - one row per flow_id
    `CREATE TABLE IF NOT EXISTS chatbot_executions_nodepath (
      id VARCHAR(255) PRIMARY KEY,
      flow_id VARCHAR(255) NOT NULL UNIQUE,
      current_node_id VARCHAR(255) NOT NULL,
      variables JSON NOT NULL,
      messages JSON NOT NULL,
      is_waiting_for_input BOOLEAN NOT NULL DEFAULT FALSE,
      is_completed BOOLEAN NOT NULL DEFAULT FALSE,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
      INDEX idx_flow_id (flow_id)
    )`,
    
    // Leads table
    `CREATE TABLE IF NOT EXISTS leads_nodepath (
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
    )`,
    
    // Media files table
    `CREATE TABLE IF NOT EXISTS media_files_nodepath (
      id VARCHAR(255) PRIMARY KEY,
      filename VARCHAR(255) NOT NULL,
      original_name VARCHAR(255) NOT NULL,
      file_type VARCHAR(255) NOT NULL,
      file_size BIGINT NOT NULL,
      storage_path VARCHAR(255) NOT NULL,
      public_url VARCHAR(255) NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    )`
  ]

  try {
    for (const sql of tables) {
      const { data: result, error } = await supabase.functions.invoke('mysql-api-bridge', {
        body: {
          endpoint: MYSQL_ENDPOINT,
          method: 'POST',
          data: { sql }
        }
      })
      
      if (error) throw error
      
      if (!result.success) {
        console.error('Failed to create table:', result.error)
      } else {
        console.log('Table created successfully')
      }
    }
  } catch (error) {
    console.error('Error initializing MySQL tables:', error)
    throw error
  }
}