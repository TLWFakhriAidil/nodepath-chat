import { ChatbotFlow, MediaFile, FlowExecution } from '@/types/chatbot'
import { supabase } from '@/integrations/supabase/client'
import { saveFlow as saveMySQLFlow, getFlows as getMySQLFlows, getFlow as getMySQLFlow, deleteFlow as deleteMySQLFlow } from './mysqlStorage'
import { saveExecution as saveMySQLExecution, getExecution as getMySQLExecution, getExecutions as getMySQLExecutions, deleteExecution as deleteMySQLExecution } from './mysqlStorage'

const MEDIA_KEY = 'chatbot_media'

// Flow management - now using Supabase
export const saveFlow = async (flow: ChatbotFlow): Promise<void> => {
  try {
    // Save main flow data to chatbot_flows table
    const { error: flowError } = await supabase
      .from('chatbot_flows')
      .upsert({
        id: flow.id,
        name: flow.name,
        description: flow.description,
        nodes: flow.nodes as any,
        edges: flow.edges as any,
        updated_at: new Date().toISOString()
      });

    if (flowError) throw flowError;

    // Extract and save AI prompt nodes separately
    const aiNodes = flow.nodes.filter(node => node.type === 'prompt');
    
    for (const node of aiNodes) {
      if (node.data.systemPrompt || node.data.instance || node.data.openRouterKey) {
        await supabase
          .from('chatbot_executions_nodepath')
          .upsert({
            id: node.id,
            system_prompt: node.data.systemPrompt || '',
            instance: node.data.instance || '',
            open_router_key: node.data.openRouterKey || '',
            conv_last: [],
            conv_current: '',
            updated_at: new Date().toISOString()
          });
      }
    }
  } catch (error) {
    console.error('Error saving flow:', error);
    throw error;
  }
}

export const getFlows = async (): Promise<ChatbotFlow[]> => {
  try {
    const { data, error } = await supabase
      .from('chatbot_flows')
      .select('*')
      .order('updated_at', { ascending: false });

    if (error) throw error;
    
    return (data || []).map(row => ({
      id: row.id,
      name: row.name,
      description: row.description || '',
      nodes: (Array.isArray(row.nodes) ? row.nodes : []) as any,
      edges: (Array.isArray(row.edges) ? row.edges : []) as any,
      createdAt: row.created_at,
      updatedAt: row.updated_at
    }));
  } catch (error) {
    console.error('Error fetching flows:', error);
    return [];
  }
}

export const getFlow = async (id: string): Promise<ChatbotFlow | null> => {
  try {
    const { data, error } = await supabase
      .from('chatbot_flows')
      .select('*')
      .eq('id', id)
      .maybeSingle();

    if (error) throw error;
    if (!data) return null;
    
    return {
      id: data.id,
      name: data.name,
      description: data.description || '',
      nodes: (Array.isArray(data.nodes) ? data.nodes : []) as any,
      edges: (Array.isArray(data.edges) ? data.edges : []) as any,
      createdAt: data.created_at,
      updatedAt: data.updated_at
    };
  } catch (error) {
    console.error('Error fetching flow:', error);
    return null;
  }
}

export const deleteFlow = async (id: string): Promise<void> => {
  return deleteMySQLFlow(id)
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

// Execution management - now using MySQL
export const saveExecution = async (execution: FlowExecution): Promise<void> => {
  return saveMySQLExecution(execution)
}

export const getExecution = async (flowId: string): Promise<FlowExecution | null> => {
  return getMySQLExecution(flowId)
}

export const getExecutions = async (): Promise<FlowExecution[]> => {
  return getMySQLExecutions()
}

export const deleteExecution = async (flowId: string): Promise<void> => {
  return deleteMySQLExecution(flowId)
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