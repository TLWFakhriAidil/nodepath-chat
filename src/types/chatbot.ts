export interface MediaFile {
  id: string
  filename: string
  type: string
  size: number
  dataUrl: string // base64 data URL for local storage
  createdAt: string
}

export interface FlowNode {
  id: string
  type: 'start' | 'message' | 'image' | 'audio' | 'video' | 'delay' | 'condition'
  position: { x: number; y: number }
  data: {
    label?: string
    message?: string
    mediaId?: string
    mediaUrl?: string
    delaySeconds?: number
    conditions?: ConditionRule[]
    variables?: Record<string, string>
  }
}

export interface ConditionRule {
  id: string
  type: 'equals' | 'contains' | 'default'
  value?: string
  nextNodeId?: string
  label: string
}

export interface FlowEdge {
  id: string
  source: string
  target: string
  sourceHandle?: string
  targetHandle?: string
}

export interface ChatbotFlow {
  id: string
  name: string
  description: string
  nodes: FlowNode[]
  edges: FlowEdge[]
  createdAt: string
  updatedAt: string
}

export interface ChatMessage {
  id: string
  type: 'user' | 'bot'
  content: string
  mediaType?: 'image' | 'audio' | 'video'
  mediaUrl?: string
  timestamp: string
}

export interface FlowExecution {
  flowId: string
  currentNodeId: string
  variables: Record<string, string>
  messages: ChatMessage[]
  isWaitingForInput: boolean
  isCompleted: boolean
}