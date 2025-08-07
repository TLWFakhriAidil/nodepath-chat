import { ChatbotFlow, MediaFile, FlowExecution } from '@/types/chatbot'
import { saveFlow as saveMySQLFlow, getFlows as getMySQLFlows, getFlow as getMySQLFlow, deleteFlow as deleteMySQLFlow } from './mysqlStorage'
import { saveFlowExecution as saveMySQLExecution, getFlowExecution as getMySQLExecution, deleteFlowExecution as deleteMySQLExecution, updateFlowExecution } from './mysqlStorage'

const MEDIA_KEY = 'chatbot_media'

// Flow management - using MySQL
export const saveFlow = async (flow: ChatbotFlow): Promise<void> => {
  return saveMySQLFlow(flow)
}

export const getFlows = async (): Promise<ChatbotFlow[]> => {
  return getMySQLFlows()
}

export const getFlow = async (id: string): Promise<ChatbotFlow | null> => {
  return getMySQLFlow(id)
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
  // For now, return empty array since we don't have a "get all executions" function
  // You can implement this if needed by creating a function in mysqlStorage.ts
  return []
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