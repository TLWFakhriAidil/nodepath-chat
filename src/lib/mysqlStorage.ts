import { ChatbotFlow, MediaFile, FlowExecution } from '@/types/chatbot'

const MYSQL_ENDPOINT = 'mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway'

// Direct MySQL API helper function - would need to be implemented with actual MySQL driver
const callMySQLAPI = async (data: any) => {
  console.warn('MySQL API bridge removed - direct MySQL connection needed')
  throw new Error('MySQL connection not available. Supabase bridge was removed.')
}

// Flow management
export const saveFlow = async (flow: ChatbotFlow): Promise<void> => {
  try {
    console.log('Saving flow to localStorage instead of MySQL:', flow.id)
    
    // Save to localStorage as fallback
    const flows = JSON.parse(localStorage.getItem('chatbot_flows') || '[]')
    const existingIndex = flows.findIndex((f: any) => f.id === flow.id)
    
    if (existingIndex >= 0) {
      flows[existingIndex] = flow
    } else {
      flows.push(flow)
    }
    
    localStorage.setItem('chatbot_flows', JSON.stringify(flows))
    console.log('Flow saved to localStorage successfully')
  } catch (error) {
    console.error('Error saving flow:', error)
    throw error
  }
}

export const getFlows = async (): Promise<ChatbotFlow[]> => {
  try {
    const flows = JSON.parse(localStorage.getItem('chatbot_flows') || '[]')
    return flows.sort((a: any, b: any) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
  } catch (error) {
    console.error('Error fetching flows:', error)
    return []
  }
}

export const getFlow = async (id: string): Promise<ChatbotFlow | null> => {
  try {
    const flows = JSON.parse(localStorage.getItem('chatbot_flows') || '[]')
    return flows.find((f: any) => f.id === id) || null
  } catch (error) {
    console.error('Error fetching flow:', error)
    return null
  }
}

export const deleteFlow = async (id: string): Promise<void> => {
  try {
    const flows = JSON.parse(localStorage.getItem('chatbot_flows') || '[]')
    const filteredFlows = flows.filter((f: any) => f.id !== id)
    localStorage.setItem('chatbot_flows', JSON.stringify(filteredFlows))
  } catch (error) {
    console.error('Error deleting flow:', error)
    throw error
  }
}

// Media management - now using localStorage
export const saveMediaFile = async (file: any): Promise<any> => {
  try {
    const mediaFiles = JSON.parse(localStorage.getItem('media_files') || '[]')
    mediaFiles.push({
      ...file,
      uploaded_at: file.uploaded_at || new Date().toISOString(),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    })
    localStorage.setItem('media_files', JSON.stringify(mediaFiles))
    return file
  } catch (error) {
    console.error('Error saving media file:', error)
    throw error
  }
}

export const getMediaFiles = async (): Promise<any[]> => {
  try {
    const mediaFiles = JSON.parse(localStorage.getItem('media_files') || '[]')
    return mediaFiles.sort((a: any, b: any) => new Date(b.uploaded_at).getTime() - new Date(a.uploaded_at).getTime())
  } catch (error) {
    console.error('Error fetching media files:', error)
    return []
  }
}

export const getMediaFile = async (id: string): Promise<any | null> => {
  try {
    const mediaFiles = JSON.parse(localStorage.getItem('media_files') || '[]')
    return mediaFiles.find((f: any) => f.id === id) || null
  } catch (error) {
    console.error('Error fetching media file:', error)
    return null
  }
}

export const deleteMediaFile = async (id: string): Promise<void> => {
  try {
    const mediaFiles = JSON.parse(localStorage.getItem('media_files') || '[]')
    const filteredFiles = mediaFiles.filter((f: any) => f.id !== id)
    localStorage.setItem('media_files', JSON.stringify(filteredFiles))
  } catch (error) {
    console.error('Error deleting media file:', error)
    throw error
  }
}

// Flow execution management - using localStorage
export const saveFlowExecution = async (execution: any): Promise<void> => {
  try {
    const simulationId = execution.id || `exec_${execution.flowId}_${Date.now()}_${Math.random().toString(36).substring(2)}`
    
    const executions = JSON.parse(localStorage.getItem('flow_executions') || '{}')
    
    executions[simulationId] = {
      ...execution,
      id: simulationId,
      updated_at: new Date().toISOString()
    }
    
    localStorage.setItem('flow_executions', JSON.stringify(executions))
    console.log('Flow execution saved to localStorage:', simulationId)
  } catch (error) {
    console.error('Error saving flow execution:', error)
    throw error
  }
}

export const getFlowExecution = async (id: string): Promise<any | null> => {
  try {
    const executions = JSON.parse(localStorage.getItem('flow_executions') || '{}')
    return executions[id] || null
  } catch (error) {
    console.error('Error fetching flow execution:', error)
    return null
  }
}

export const updateFlowExecution = async (id: string, updates: any): Promise<void> => {
  try {
    const executions = JSON.parse(localStorage.getItem('flow_executions') || '{}')
    
    if (executions[id]) {
      executions[id] = {
        ...executions[id],
        ...updates,
        updated_at: new Date().toISOString()
      }
      localStorage.setItem('flow_executions', JSON.stringify(executions))
    }
  } catch (error) {
    console.error('Error updating flow execution:', error)
    throw error
  }
}

export const deleteFlowExecution = async (id: string): Promise<void> => {
  try {
    const executions = JSON.parse(localStorage.getItem('flow_executions') || '{}')
    delete executions[id]
    localStorage.setItem('flow_executions', JSON.stringify(executions))
  } catch (error) {
    console.error('Error deleting flow execution:', error)
    throw error
  }
}

// Get AI settings - using localStorage
export const getAISettings = async () => {
  try {
    const settings = JSON.parse(localStorage.getItem('ai_settings') || 'null')
    return settings
  } catch (error) {
    console.error('Error fetching AI settings:', error)
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