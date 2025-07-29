import { ChatbotFlow, MediaFile, FlowExecution } from '@/types/chatbot'

const FLOWS_KEY = 'chatbot_flows'
const MEDIA_KEY = 'chatbot_media'
const EXECUTIONS_KEY = 'chatbot_executions'

// Flow management
export const saveFlow = (flow: ChatbotFlow): void => {
  const flows = getFlows()
  const existingIndex = flows.findIndex(f => f.id === flow.id)
  
  if (existingIndex >= 0) {
    flows[existingIndex] = { ...flow, updatedAt: new Date().toISOString() }
  } else {
    flows.push(flow)
  }
  
  localStorage.setItem(FLOWS_KEY, JSON.stringify(flows))
}

export const getFlows = (): ChatbotFlow[] => {
  const stored = localStorage.getItem(FLOWS_KEY)
  return stored ? JSON.parse(stored) : []
}

export const getFlow = (id: string): ChatbotFlow | null => {
  const flows = getFlows()
  return flows.find(f => f.id === id) || null
}

export const deleteFlow = (id: string): void => {
  const flows = getFlows().filter(f => f.id !== id)
  localStorage.setItem(FLOWS_KEY, JSON.stringify(flows))
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
export const saveExecution = (execution: FlowExecution): void => {
  const executions = getExecutions()
  const existingIndex = executions.findIndex(e => e.flowId === execution.flowId)
  
  if (existingIndex >= 0) {
    executions[existingIndex] = execution
  } else {
    executions.push(execution)
  }
  
  localStorage.setItem(EXECUTIONS_KEY, JSON.stringify(executions))
}

export const getExecution = (flowId: string): FlowExecution | null => {
  const executions = getExecutions()
  return executions.find(e => e.flowId === flowId) || null
}

export const getExecutions = (): FlowExecution[] => {
  const stored = localStorage.getItem(EXECUTIONS_KEY)
  return stored ? JSON.parse(stored) : []
}

export const deleteExecution = (flowId: string): void => {
  const executions = getExecutions().filter(e => e.flowId !== flowId)
  localStorage.setItem(EXECUTIONS_KEY, JSON.stringify(executions))
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