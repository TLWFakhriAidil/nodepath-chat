import { ChatbotFlow, FlowNode, FlowExecution, ChatMessage, ConditionRule } from '@/types/chatbot'
import { getFlow, getMediaFile, saveExecution, replaceVariables } from '@/lib/localStorage'

export class FlowEngine {
  private execution: FlowExecution
  private flow: ChatbotFlow
  private flowId: string
  private onMessage: (message: ChatMessage) => void
  private onComplete: () => void
  private onWaitingForInput: () => void

  constructor(
    flowId: string, 
    onMessage: (message: ChatMessage) => void,
    onComplete: () => void,
    onWaitingForInput: () => void
  ) {
    this.flowId = flowId
    this.onMessage = onMessage
    this.onComplete = onComplete
    this.onWaitingForInput = onWaitingForInput
  }

  async initialize() {
    const flow = await getFlow(this.flowId)
    if (!flow) {
      throw new Error(`Flow with id ${this.flowId} not found`)
    }

    this.flow = flow

    // Initialize execution
    const startNode = flow.nodes.find(node => node.type === 'start')
    console.log('FlowEngine constructor - Start node found:', startNode)
    console.log('FlowEngine constructor - All nodes:', flow.nodes)
    console.log('FlowEngine constructor - All edges:', flow.edges)
    
    if (!startNode) {
      throw new Error('Flow must have a start node')
    }

    this.execution = {
      flowId: this.flowId,
      currentNodeId: startNode.id,
      variables: { username: 'User' }, // Default variables
      messages: [],
      isWaitingForInput: false,
      isCompleted: false
    }

    console.log('FlowEngine constructor - Initial execution state:', this.execution)
  }

  async start(): Promise<void> {
    console.log('FlowEngine.start() called')
    console.log('Current execution state:', this.execution)
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
    console.log('Looking for next node from:', sourceId)
    console.log('Available edges:', this.flow.edges)
    
    const edge = this.flow.edges.find(edge => edge.source === sourceId)
    console.log('Found edge:', edge)
    
    if (!edge) {
      console.log('No edge found from node:', sourceId)
      return undefined
    }
    
    const nextNode = this.flow.nodes.find(node => node.id === edge.target)
    console.log('Next node:', nextNode)
    return nextNode
  }

  private async processCurrentNode(): Promise<void> {
    const currentNode = this.getCurrentNode()
    console.log('Processing current node:', currentNode)
    
    if (!currentNode) {
      console.log('No current node found, completing execution')
      await this.completeExecution()
      return
    }

    console.log('Node type:', currentNode.type)
    
    switch (currentNode.type) {
      case 'start':
        console.log('Processing start node - moving to next')
        await this.moveToNextNode()
        break
      
      case 'message':
        console.log('Processing message node')
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

    await this.saveState()
  }

  private async handleMessageNode(node: FlowNode): Promise<void> {
    console.log('handleMessageNode called with:', node)
    const message = replaceVariables(node.data.message || '', this.execution.variables)
    console.log('Message to send:', message)
    
    const botMessage: ChatMessage = {
      id: `msg_${Date.now()}_${Math.random().toString(36).substring(2)}`,
      type: 'bot',
      content: message,
      timestamp: new Date().toISOString()
    }

    console.log('Bot message created:', botMessage)
    this.execution.messages.push(botMessage)
    this.onMessage(botMessage)

    await this.moveToNextNode()
  }

  private async handleImageNode(node: FlowNode): Promise<void> {
    console.log('Image node data:', node.data)
    let mediaUrl = node.data.mediaUrl || node.data.imageUrl

    // Try previewUrl if uploaded file exists
    if (node.data.previewUrl) {
      mediaUrl = node.data.previewUrl
      console.log('Using previewUrl:', mediaUrl)
    }

    if (node.data.mediaId) {
      const mediaFile = getMediaFile(node.data.mediaId)
      if (mediaFile) {
        mediaUrl = mediaFile.dataUrl
        console.log('Using mediaFile.dataUrl:', mediaUrl)
      }
    }

    console.log('Final mediaUrl for image:', mediaUrl)

    if (mediaUrl) {
      const botMessage: ChatMessage = {
        id: `msg_${Date.now()}_${Math.random().toString(36).substring(2)}`,
        type: 'bot',
        content: node.data.caption || node.data.message || '',
        mediaType: 'image',
        mediaUrl,
        timestamp: new Date().toISOString()
      }

      console.log('Sending image message:', botMessage)
      this.execution.messages.push(botMessage)
      this.onMessage(botMessage)
    } else {
      console.log('No mediaUrl found for image node')
    }

    await this.moveToNextNode()
  }

  private async handleAudioNode(node: FlowNode): Promise<void> {
    console.log('Audio node data:', node.data)
    let mediaUrl = node.data.mediaUrl || node.data.audioUrl

    // Try previewUrl if uploaded file exists
    if (node.data.previewUrl) {
      mediaUrl = node.data.previewUrl
      console.log('Using previewUrl:', mediaUrl)
    }

    if (node.data.mediaId) {
      const mediaFile = getMediaFile(node.data.mediaId)
      if (mediaFile) {
        mediaUrl = mediaFile.dataUrl
        console.log('Using mediaFile.dataUrl:', mediaUrl)
      }
    }

    console.log('Final mediaUrl for audio:', mediaUrl)

    if (mediaUrl) {
      const botMessage: ChatMessage = {
        id: `msg_${Date.now()}_${Math.random().toString(36).substring(2)}`,
        type: 'bot',
        content: node.data.message || '',
        mediaType: 'audio',
        mediaUrl,
        timestamp: new Date().toISOString()
      }

      console.log('Sending audio message:', botMessage)
      this.execution.messages.push(botMessage)
      this.onMessage(botMessage)
    } else {
      console.log('No mediaUrl found for audio node')
    }

    await this.moveToNextNode()
  }

  private async handleVideoNode(node: FlowNode): Promise<void> {
    console.log('Video node data:', node.data)
    let mediaUrl = node.data.mediaUrl || node.data.videoUrl

    // Try previewUrl if uploaded file exists
    if (node.data.previewUrl) {
      mediaUrl = node.data.previewUrl
      console.log('Using previewUrl:', mediaUrl)
    }

    if (node.data.mediaId) {
      const mediaFile = getMediaFile(node.data.mediaId)
      if (mediaFile) {
        mediaUrl = mediaFile.dataUrl
        console.log('Using mediaFile.dataUrl:', mediaUrl)
      }
    }

    console.log('Final mediaUrl for video:', mediaUrl)

    if (mediaUrl) {
      const botMessage: ChatMessage = {
        id: `msg_${Date.now()}_${Math.random().toString(36).substring(2)}`,
        type: 'bot',
        content: node.data.caption || node.data.message || '',
        mediaType: 'video',
        mediaUrl,
        timestamp: new Date().toISOString()
      }

      console.log('Sending video message:', botMessage)
      this.execution.messages.push(botMessage)
      this.onMessage(botMessage)
    } else {
      console.log('No mediaUrl found for video node')
    }

    await this.moveToNextNode()
  }

  private async handleDelayNode(node: FlowNode): Promise<void> {
    const delaySeconds = node.data.delaySeconds || node.data.delay || 1
    
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

    if (matchedCondition) {
      // For condition nodes, find the next node based on the outgoing edge with the matching handle
      const conditionEdge = this.flow.edges.find(edge => 
        edge.source === node.id && edge.sourceHandle === matchedCondition!.id
      )
      
      if (conditionEdge) {
        this.execution.currentNodeId = conditionEdge.target
        await this.processCurrentNode()
      } else {
        console.log('No edge found for condition:', matchedCondition)
        await this.moveToNextNode()
      }
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
      await this.completeExecution()
    }
  }

  private async completeExecution(): Promise<void> {
    this.execution.isCompleted = true
    this.execution.isWaitingForInput = false
    await this.saveState()
    this.onComplete()
  }

  private async saveState(): Promise<void> {
    await saveExecution(this.execution)
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