import { ChatbotFlow, FlowNode, FlowExecution, ChatMessage, ConditionRule } from '@/types/chatbot'
import { getFlow, getMediaFile, saveExecution, replaceVariables } from '@/lib/localStorage'

export class FlowEngine {
  private execution: FlowExecution
  private flow: ChatbotFlow
  private onMessage: (message: ChatMessage) => void
  private onComplete: () => void
  private onWaitingForInput: () => void

  constructor(
    flowId: string,
    onMessage: (message: ChatMessage) => void,
    onComplete: () => void,
    onWaitingForInput: () => void
  ) {
    const flow = getFlow(flowId)
    if (!flow) {
      throw new Error(`Flow with id ${flowId} not found`)
    }

    this.flow = flow
    this.onMessage = onMessage
    this.onComplete = onComplete
    this.onWaitingForInput = onWaitingForInput

    // Initialize execution
    const startNode = flow.nodes.find(node => node.type === 'start')
    if (!startNode) {
      throw new Error('Flow must have a start node')
    }

    this.execution = {
      flowId,
      currentNodeId: startNode.id,
      variables: { username: 'User' }, // Default variables
      messages: [],
      isWaitingForInput: false,
      isCompleted: false
    }
  }

  async start(): Promise<void> {
    await this.processCurrentNode()
  }

  async processUserInput(input: string): Promise<void> {
    if (!this.execution.isWaitingForInput) {
      return
    }

    // Add user message
    const userMessage: ChatMessage = {
      id: `msg_${Date.now()}_${Math.random().toString(36).substring(2)}`,
      type: 'user',
      content: input,
      timestamp: new Date().toISOString()
    }

    this.execution.messages.push(userMessage)
    this.onMessage(userMessage)

    // Update variables if needed
    this.execution.variables.lastInput = input

    this.execution.isWaitingForInput = false

    // Process condition node if current node is condition
    const currentNode = this.getCurrentNode()
    if (currentNode?.type === 'condition') {
      await this.handleConditionNode(currentNode, input)
    } else {
      await this.moveToNextNode()
    }
  }

  private getCurrentNode(): FlowNode | undefined {
    return this.flow.nodes.find(node => node.id === this.execution.currentNodeId)
  }

  private getNextNode(fromNodeId?: string): FlowNode | undefined {
    const sourceId = fromNodeId || this.execution.currentNodeId
    const edge = this.flow.edges.find(edge => edge.source === sourceId)
    if (!edge) return undefined
    
    return this.flow.nodes.find(node => node.id === edge.target)
  }

  private async processCurrentNode(): Promise<void> {
    const currentNode = this.getCurrentNode()
    if (!currentNode) {
      this.completeExecution()
      return
    }

    switch (currentNode.type) {
      case 'start':
        await this.moveToNextNode()
        break
      
      case 'message':
        await this.handleMessageNode(currentNode)
        break
      
      case 'image':
        await this.handleImageNode(currentNode)
        break
      
      case 'audio':
        await this.handleAudioNode(currentNode)
        break
      
      case 'video':
        await this.handleVideoNode(currentNode)
        break
      
      case 'delay':
        await this.handleDelayNode(currentNode)
        break
      
      case 'condition':
        await this.handleConditionNode(currentNode)
        break
      
      default:
        console.warn(`Unknown node type: ${currentNode.type}`)
        await this.moveToNextNode()
    }

    this.saveState()
  }

  private async handleMessageNode(node: FlowNode): Promise<void> {
    const message = replaceVariables(node.data.message || '', this.execution.variables)
    
    const botMessage: ChatMessage = {
      id: `msg_${Date.now()}_${Math.random().toString(36).substring(2)}`,
      type: 'bot',
      content: message,
      timestamp: new Date().toISOString()
    }

    this.execution.messages.push(botMessage)
    this.onMessage(botMessage)

    await this.moveToNextNode()
  }

  private async handleImageNode(node: FlowNode): Promise<void> {
    let mediaUrl = node.data.mediaUrl

    if (node.data.mediaId) {
      const mediaFile = getMediaFile(node.data.mediaId)
      if (mediaFile) {
        mediaUrl = mediaFile.dataUrl
      }
    }

    if (mediaUrl) {
      const botMessage: ChatMessage = {
        id: `msg_${Date.now()}_${Math.random().toString(36).substring(2)}`,
        type: 'bot',
        content: node.data.message || '',
        mediaType: 'image',
        mediaUrl,
        timestamp: new Date().toISOString()
      }

      this.execution.messages.push(botMessage)
      this.onMessage(botMessage)
    }

    await this.moveToNextNode()
  }

  private async handleAudioNode(node: FlowNode): Promise<void> {
    let mediaUrl = node.data.mediaUrl

    if (node.data.mediaId) {
      const mediaFile = getMediaFile(node.data.mediaId)
      if (mediaFile) {
        mediaUrl = mediaFile.dataUrl
      }
    }

    if (mediaUrl) {
      const botMessage: ChatMessage = {
        id: `msg_${Date.now()}_${Math.random().toString(36).substring(2)}`,
        type: 'bot',
        content: node.data.message || '',
        mediaType: 'audio',
        mediaUrl,
        timestamp: new Date().toISOString()
      }

      this.execution.messages.push(botMessage)
      this.onMessage(botMessage)
    }

    await this.moveToNextNode()
  }

  private async handleVideoNode(node: FlowNode): Promise<void> {
    let mediaUrl = node.data.mediaUrl

    if (node.data.mediaId) {
      const mediaFile = getMediaFile(node.data.mediaId)
      if (mediaFile) {
        mediaUrl = mediaFile.dataUrl
      }
    }

    if (mediaUrl) {
      const botMessage: ChatMessage = {
        id: `msg_${Date.now()}_${Math.random().toString(36).substring(2)}`,
        type: 'bot',
        content: node.data.message || '',
        mediaType: 'video',
        mediaUrl,
        timestamp: new Date().toISOString()
      }

      this.execution.messages.push(botMessage)
      this.onMessage(botMessage)
    }

    await this.moveToNextNode()
  }

  private async handleDelayNode(node: FlowNode): Promise<void> {
    const delaySeconds = node.data.delaySeconds || 1
    
    setTimeout(async () => {
      await this.moveToNextNode()
    }, delaySeconds * 1000)
  }

  private async handleConditionNode(node: FlowNode, input?: string): Promise<void> {
    if (!input && !this.execution.isWaitingForInput) {
      // First time reaching condition node, wait for user input
      this.execution.isWaitingForInput = true
      this.onWaitingForInput()
      return
    }

    const userInput = input || this.execution.variables.lastInput || ''
    const conditions = node.data.conditions || []

    // Find matching condition
    let matchedCondition: ConditionRule | undefined

    for (const condition of conditions) {
      if (condition.type === 'equals' && userInput.toLowerCase() === (condition.value || '').toLowerCase()) {
        matchedCondition = condition
        break
      } else if (condition.type === 'contains' && userInput.toLowerCase().includes((condition.value || '').toLowerCase())) {
        matchedCondition = condition
        break
      }
    }

    // Use default condition if no match
    if (!matchedCondition) {
      matchedCondition = conditions.find(c => c.type === 'default')
    }

    if (matchedCondition?.nextNodeId) {
      this.execution.currentNodeId = matchedCondition.nextNodeId
      await this.processCurrentNode()
    } else {
      await this.moveToNextNode()
    }
  }

  private async moveToNextNode(): Promise<void> {
    const nextNode = this.getNextNode()
    
    if (nextNode) {
      this.execution.currentNodeId = nextNode.id
      await this.processCurrentNode()
    } else {
      this.completeExecution()
    }
  }

  private completeExecution(): void {
    this.execution.isCompleted = true
    this.execution.isWaitingForInput = false
    this.saveState()
    this.onComplete()
  }

  private saveState(): void {
    saveExecution(this.execution)
  }

  getMessages(): ChatMessage[] {
    return this.execution.messages
  }

  isWaitingForInput(): boolean {
    return this.execution.isWaitingForInput
  }

  isCompleted(): boolean {
    return this.execution.isCompleted
  }
}