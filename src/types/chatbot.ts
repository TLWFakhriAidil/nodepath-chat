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
  type: 'start' | 'message' | 'image' | 'audio' | 'video' | 'delay' | 'condition' | 'manual' | 'prompt' | 'stage' | 'user_reply' | 'waiting_reply_times'
  position: { x: number; y: number }
  data: {
    label?: string
    message?: string
    mediaId?: string
    mediaUrl?: string
    imageUrl?: string
    audioUrl?: string
    videoUrl?: string

    duration?: number
    previewUrl?: string
    uploadedFile?: {
      name: string
      type: string
      size: number
    } | null
    delay?: number
    delaySeconds?: number
    waitTime?: number
    waitTimeSeconds?: number
    conditions?: ConditionRule[]
    variables?: Record<string, string>
    // Manual node fields
    expectedInput?: string
    responseOutput?: string
    // AI Prompt node fields
    systemPrompt?: string
    instance?: string
    openRouterKey?: string
    node_type?: string
    // Stage node fields
    stageName?: string
    // Node update function
    onUpdate?: (nodeId: string, data: any) => void
    onDelete?: (nodeId: string) => void
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
  niche?: string
  selectedDeviceId?: string
  id_device?: string
  idDevice?: string
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
  id?: string
  flowId: string
  currentNodeId: string
  variables: Record<string, string>
  messages: ChatMessage[]
  isWaitingForInput: boolean
  isCompleted: boolean
}