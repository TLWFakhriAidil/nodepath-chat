import { ChatbotFlow, MediaFile, FlowExecution } from '@/types/chatbot'
import { supabase } from '@/integrations/supabase/client'

const MEDIA_KEY = 'chatbot_media'

// Flow management
export const saveFlow = async (flow: ChatbotFlow): Promise<void> => {
  try {
    const { error } = await supabase
      .from('chatbot_flows')
      .upsert({
        id: flow.id,
        name: flow.name,
        description: flow.description,
        nodes: flow.nodes as any,
        edges: flow.edges as any,
        updated_at: new Date().toISOString()
      })
    
    if (error) throw error
  } catch (error) {
    console.error('Error saving flow:', error)
    throw error
  }
}

export const getFlows = async (): Promise<ChatbotFlow[]> => {
  try {
    const { data, error } = await supabase
      .from('chatbot_flows')
      .select('*')
      .order('updated_at', { ascending: false })
    
    if (error) throw error
    
    return (data || []).map(row => ({
      id: row.id,
      name: row.name,
      description: row.description,
      nodes: row.nodes as any,
      edges: row.edges as any,
      createdAt: row.created_at,
      updatedAt: row.updated_at
    }))
  } catch (error) {
    console.error('Error fetching flows:', error)
    return []
  }
}

export const getFlow = async (id: string): Promise<ChatbotFlow | null> => {
  try {
    const { data, error } = await supabase
      .from('chatbot_flows')
      .select('*')
      .eq('id', id)
      .maybeSingle()
    
    if (error) throw error
    if (!data) return null
    
    return {
      id: data.id,
      name: data.name,
      description: data.description,
      nodes: data.nodes as any,
      edges: data.edges as any,
      createdAt: data.created_at,
      updatedAt: data.updated_at
    }
  } catch (error) {
    console.error('Error fetching flow:', error)
    return null
  }
}

export const deleteFlow = async (id: string): Promise<void> => {
  try {
    const { error } = await supabase
      .from('chatbot_flows')
      .delete()
      .eq('id', id)
    
    if (error) throw error
  } catch (error) {
    console.error('Error deleting flow:', error)
    throw error
  }
}

// Media management
export const saveMediaFile = (file: MediaFile): void => {
  const media = getMediaFiles()
  media.push(file)
  localStorage.setItem(MEDIA_KEY, JSON.stringify(media))
}

export const getMediaFiles = (): MediaFile[] => {
  const stored = localStorage.getItem(MEDIA_KEY)
  return stored ? JSON.parse(stored) : []
}

export const getMediaFile = (id: string): MediaFile | null => {
  const media = getMediaFiles()
  return media.find(m => m.id === id) || null
}

export const deleteMediaFile = (id: string): void => {
  const media = getMediaFiles().filter(m => m.id !== id)
  localStorage.setItem(MEDIA_KEY, JSON.stringify(media))
}

// Execution management
export const saveExecution = async (execution: FlowExecution): Promise<void> => {
  try {
    const { error } = await supabase
      .from('chatbot_executions')
      .upsert({
        flow_id: execution.flowId,
        current_node_id: execution.currentNodeId,
        variables: execution.variables as any,
        messages: execution.messages as any,
        is_waiting_for_input: execution.isWaitingForInput,
        is_completed: execution.isCompleted
      })
    
    if (error) throw error
  } catch (error) {
    console.error('Error saving execution:', error)
    throw error
  }
}

export const getExecution = async (flowId: string): Promise<FlowExecution | null> => {
  try {
    const { data, error } = await supabase
      .from('chatbot_executions')
      .select('*')
      .eq('flow_id', flowId)
      .maybeSingle()
    
    if (error) throw error
    if (!data) return null
    
    return {
      flowId: data.flow_id,
      currentNodeId: data.current_node_id,
      variables: data.variables as any,
      messages: data.messages as any,
      isWaitingForInput: data.is_waiting_for_input,
      isCompleted: data.is_completed
    }
  } catch (error) {
    console.error('Error fetching execution:', error)
    return null
  }
}

export const getExecutions = async (): Promise<FlowExecution[]> => {
  try {
    const { data, error } = await supabase
      .from('chatbot_executions')
      .select('*')
      .order('updated_at', { ascending: false })
    
    if (error) throw error
    
    return (data || []).map(row => ({
      flowId: row.flow_id,
      currentNodeId: row.current_node_id,
      variables: row.variables as any,
      messages: row.messages as any,
      isWaitingForInput: row.is_waiting_for_input,
      isCompleted: row.is_completed
    }))
  } catch (error) {
    console.error('Error fetching executions:', error)
    return []
  }
}

export const deleteExecution = async (flowId: string): Promise<void> => {
  try {
    const { error } = await supabase
      .from('chatbot_executions')
      .delete()
      .eq('flow_id', flowId)
    
    if (error) throw error
  } catch (error) {
    console.error('Error deleting execution:', error)
    throw error
  }
}

// Utility functions
export const createMediaFileFromFile = async (file: File): Promise<MediaFile> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    
    reader.onload = (e) => {
      const dataUrl = e.target?.result as string
      
      const mediaFile: MediaFile = {
        id: `media_${Date.now()}_${Math.random().toString(36).substring(2)}`,
        filename: file.name,
        type: file.type,
        size: file.size,
        dataUrl,
        createdAt: new Date().toISOString()
      }
      
      resolve(mediaFile)
    }
    
    reader.onerror = () => reject(new Error('Failed to read file'))
    reader.readAsDataURL(file)
  })
}

export const replaceVariables = (text: string, variables: Record<string, string>): string => {
  let result = text
  Object.entries(variables).forEach(([key, value]) => {
    result = result.replace(new RegExp(`{{${key}}}`, 'g'), value)
  })
  return result
}