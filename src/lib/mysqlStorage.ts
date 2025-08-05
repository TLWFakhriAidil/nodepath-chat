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
    console.log('Saving flow to MySQL:', flow.id)
    
    // Check if flow already exists
    const existingResult = await callMySQLAPI({
      operation: 'select',
      table: 'chatbot_flows_nodepath',
      filters: { id: flow.id }
    })
    
    const flowExists = existingResult.success && existingResult.data && existingResult.data.length > 0
    console.log('Flow exists:', flowExists)

    // Extract AI prompt data from nodes
    const aiPromptNode = flow.nodes.find(node => node.type === 'prompt')
    const aiPromptData = aiPromptNode ? {
      instance: aiPromptNode.data.instance || '',
      open_router_key: aiPromptNode.data.openRouterKey || ''
    } : {
      instance: '',
      open_router_key: ''
    }

    // Prepare flow data for MySQL storage (including AI prompt data as separate columns)
    const flowData = {
      id: flow.id,
      name: flow.name,
      description: flow.description,
      nodes: JSON.stringify(flow.nodes),
      edges: JSON.stringify(flow.edges),
      created_at: flow.createdAt,
      updated_at: flow.updatedAt,
      ...aiPromptData  // Add AI prompt data as separate columns
    }

    // Save all flow data to chatbot_flows_nodepath
    const saveResult = await callMySQLAPI({
      operation: flowExists ? 'update' : 'insert',
      table: 'chatbot_flows_nodepath',
      ...(flowExists ? { filters: { id: flow.id } } : {}),
      payload: flowData
    })

    if (!saveResult.success) {
      throw new Error(`Failed to save flow: ${saveResult.error}`)
    }

    console.log('Flow saved to MySQL successfully with AI prompt data in separate columns')
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

// Media management
export const saveMediaFile = async (file: any): Promise<any> => {
  try {
    const result = await callMySQLAPI({
      operation: 'insert',
      table: 'media_files',
      payload: {
        id: file.id,
        filename: file.filename,
        original_name: file.original_name || file.originalName,
        file_type: file.file_type || file.fileType,
        file_size: file.file_size || file.fileSize,
        storage_path: file.storage_path || file.storagePath,
        public_url: file.public_url || file.publicUrl,
        uploaded_at: file.uploaded_at || file.uploadedAt || new Date().toISOString(),
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
    })
    
    if (!result.success) throw new Error(result.error)
    return file
  } catch (error) {
    console.error('Error saving media file to MySQL:', error)
    throw error
  }
}

export const getMediaFiles = async (): Promise<any[]> => {
  try {
    const result = await callMySQLAPI({
      operation: 'select',
      table: 'media_files',
      filters: { order_by: 'uploaded_at DESC' }
    })
    
    if (!result.success) throw new Error(result.error)
    
    return (result.data || []).map((row: any) => ({
      id: row.id,
      filename: row.filename,
      original_name: row.original_name,
      file_type: row.file_type,
      file_size: row.file_size,
      storage_path: row.storage_path,
      public_url: row.public_url,
      uploaded_at: row.uploaded_at,
      created_at: row.created_at,
      updated_at: row.updated_at
    }))
  } catch (error) {
    console.error('Error fetching media files from MySQL:', error)
    return []
  }
}

export const getMediaFile = async (id: string): Promise<any | null> => {
  try {
    const result = await callMySQLAPI({
      operation: 'select',
      table: 'media_files',
      filters: { id }
    })
    
    if (!result.success) throw new Error(result.error)
    if (!result.data || result.data.length === 0) return null
    
    const row = result.data[0]
    return {
      id: row.id,
      filename: row.filename,
      original_name: row.original_name,
      file_type: row.file_type,
      file_size: row.file_size,
      storage_path: row.storage_path,
      public_url: row.public_url,
      uploaded_at: row.uploaded_at,
      created_at: row.created_at,
      updated_at: row.updated_at
    }
  } catch (error) {
    console.error('Error fetching media file from MySQL:', error)
    return null
  }
}

export const deleteMediaFile = async (id: string): Promise<void> => {
  try {
    const result = await callMySQLAPI({
      operation: 'delete',
      table: 'media_files',
      filters: { id }
    })
    
    if (!result.success) throw new Error(result.error)
  } catch (error) {
    console.error('Error deleting media file from MySQL:', error)
    throw error
  }
}

// Flow execution management
export const saveFlowExecution = async (execution: any): Promise<void> => {
  try {
    // Create a unique simulation ID based on flow and start time if not provided
    const simulationId = execution.id || `exec_${execution.flowId}_${Date.now()}_${Math.random().toString(36).substring(2)}`
    
    // Check if execution already exists for this simulation
    const existingExecution = await getFlowExecution(simulationId)
    
    // Get instance from flow data
    const flowData = await getFlow(execution.flowId)
    const aiPromptNode = flowData?.nodes?.find(node => node.type === 'prompt')
    const instance = aiPromptNode?.data?.instance || ''

    const payload: any = {
      id: simulationId,
      flow_id: execution.flowId,
      current_node_id: execution.currentNodeId,
      variables: JSON.stringify(execution.variables || {}),
      messages: JSON.stringify(execution.messages || []),
      is_waiting_for_input: execution.isWaitingForInput || false,
      is_completed: execution.isCompleted || false,
      instance: instance,
      updated_at: new Date().toISOString()
    }

    if (existingExecution) {
      // Update existing execution with complete conversation
      const result = await callMySQLAPI({
        operation: 'update',
        table: 'chatbot_executions_nodepath',
        filters: { id: simulationId },
        payload
      })
      
      if (!result.success) throw new Error(result.error)
      console.log('Flow execution updated successfully for simulation:', simulationId)
    } else {
      // Create new execution for this simulation
      payload.created_at = new Date().toISOString()
      
      const result = await callMySQLAPI({
        operation: 'insert',
        table: 'chatbot_executions_nodepath',
        payload
      })
      
      if (!result.success) throw new Error(result.error)
      console.log('Flow execution created successfully for simulation:', simulationId)
    }
  } catch (error) {
    console.error('Error saving flow execution to MySQL:', error)
    throw error
  }
}

export const getFlowExecution = async (id: string): Promise<any | null> => {
  try {
    const result = await callMySQLAPI({
      operation: 'select',
      table: 'chatbot_executions_nodepath',
      filters: { id }
    })
    
    if (!result.success) throw new Error(result.error)
    if (!result.data || result.data.length === 0) return null
    
    const row = result.data[0]
    return {
      id: row.id,
      flowId: row.flow_id,
      currentNodeId: row.current_node_id,
      variables: JSON.parse(row.variables || '{}'),
      messages: JSON.parse(row.messages || '[]'),
      isWaitingForInput: row.is_waiting_for_input || false,
      isCompleted: row.is_completed || false,
      createdAt: row.created_at,
      updatedAt: row.updated_at
    }
  } catch (error) {
    console.error('Error fetching flow execution from MySQL:', error)
    return null
  }
}

export const updateFlowExecution = async (id: string, updates: any): Promise<void> => {
  try {
    const payload: any = {}
    
    if (updates.currentNodeId !== undefined) payload.current_node_id = updates.currentNodeId
    if (updates.variables !== undefined) payload.variables = JSON.stringify(updates.variables)
    if (updates.messages !== undefined) payload.messages = JSON.stringify(updates.messages)
    if (updates.isWaitingForInput !== undefined) payload.is_waiting_for_input = updates.isWaitingForInput
    if (updates.isCompleted !== undefined) payload.is_completed = updates.isCompleted
    
    // Always update the timestamp
    payload.updated_at = new Date().toISOString()
    
    const result = await callMySQLAPI({
      operation: 'update',
      table: 'chatbot_executions_nodepath',
      filters: { id },
      payload
    })
    
    if (!result.success) throw new Error(result.error)
  } catch (error) {
    console.error('Error updating flow execution in MySQL:', error)
    throw error
  }
}

export const deleteFlowExecution = async (id: string): Promise<void> => {
  try {
    const result = await callMySQLAPI({
      operation: 'delete',
      table: 'chatbot_executions_nodepath',
      filters: { id }
    })
    
    if (!result.success) throw new Error(result.error)
  } catch (error) {
    console.error('Error deleting flow execution from MySQL:', error)
    throw error
  }
}

// Get AI settings for chat execution
export const getAISettings = async () => {
  try {
    const result = await callMySQLAPI({
      sql: 'SELECT * FROM ai_settings_nodepath ORDER BY created_at DESC LIMIT 1'
    })
    
    if (!result.success) throw new Error(result.error)
    if (!result.data?.result?.rows || result.data.result.rows.length === 0) {
      return null
    }
    
    return result.data.result.rows[0]
  } catch (error) {
    console.error('Error fetching AI settings from MySQL:', error)
    return null
  }
}

// Helper function to extract AI prompt data from flow nodes
export const extractAIPromptData = (flow: ChatbotFlow) => {
  const aiNodes = flow.nodes.filter(node => node.type === 'prompt')
  return aiNodes.map(node => ({
    nodeId: node.id,
    instance: node.data.instance || '',
    openRouterKey: node.data.openRouterKey || ''
  }))
}