import { ChatbotFlow, FlowNode, FlowExecution, ChatMessage, ConditionRule } from '@/types/chatbot'
import { getFlow, getMediaFile, saveExecution, replaceVariables } from '@/lib/localStorage'

export class FlowEngine {
  private execution: FlowExecution
  private flow: ChatbotFlow
  private flowId: string
  private simulationId: string
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
    this.simulationId = `exec_${flowId}_${Date.now()}_${Math.random().toString(36).substring(2)}`
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
    } else if (currentNode?.type === 'prompt') {
      // Handle AI prompt node with user input
      await this.handlePromptNode(currentNode)
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

      case 'prompt':
        await this.handlePromptNode(currentNode)
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

    // Handle wrapped string objects from MySQL storage
    if (mediaUrl && typeof mediaUrl === 'object' && (mediaUrl as any)._type === 'String') {
      mediaUrl = (mediaUrl as any).value
    }

    // Try previewUrl if uploaded file exists
    if (node.data.previewUrl) {
      let previewUrl = node.data.previewUrl
      if (previewUrl && typeof previewUrl === 'object' && (previewUrl as any)._type === 'String') {
        previewUrl = (previewUrl as any).value
      }
      mediaUrl = previewUrl
      console.log('Using previewUrl:', mediaUrl ? `${mediaUrl.substring(0, 50)}...` : 'null')
    }

    if (node.data.mediaId) {
      const mediaFile = getMediaFile(node.data.mediaId)
      if (mediaFile) {
        mediaUrl = mediaFile.dataUrl
        console.log('Using mediaFile.dataUrl:', mediaUrl ? `${mediaUrl.substring(0, 50)}...` : 'null')
      }
    }

    console.log('Final mediaUrl for image:', mediaUrl ? `${mediaUrl.substring(0, 50)}...` : 'null')
    console.log('MediaUrl is valid base64?', mediaUrl ? mediaUrl.startsWith('data:image/') : false)

    if (mediaUrl) {
      const botMessage: ChatMessage = {
        id: `msg_${Date.now()}_${Math.random().toString(36).substring(2)}`,
        type: 'bot',
        content: node.data.message || '',
        mediaType: 'image',
        mediaUrl,
        timestamp: new Date().toISOString()
      }

      console.log('Sending image message with mediaUrl length:', mediaUrl.length)
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

    // Handle wrapped string objects from MySQL storage
    if (mediaUrl && typeof mediaUrl === 'object' && (mediaUrl as any)._type === 'String') {
      mediaUrl = (mediaUrl as any).value
    }

    // Try previewUrl if uploaded file exists
    if (node.data.previewUrl) {
      let previewUrl = node.data.previewUrl
      if (previewUrl && typeof previewUrl === 'object' && (previewUrl as any)._type === 'String') {
        previewUrl = (previewUrl as any).value
      }
      mediaUrl = previewUrl
      console.log('Using previewUrl:', mediaUrl ? `${mediaUrl.substring(0, 50)}...` : 'null')
    }

    if (node.data.mediaId) {
      const mediaFile = getMediaFile(node.data.mediaId)
      if (mediaFile) {
        mediaUrl = mediaFile.dataUrl
        console.log('Using mediaFile.dataUrl:', mediaUrl ? `${mediaUrl.substring(0, 50)}...` : 'null')
      }
    }

    console.log('Final mediaUrl for audio:', mediaUrl ? `${mediaUrl.substring(0, 50)}...` : 'null')

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

    // Handle wrapped string objects from MySQL storage
    if (mediaUrl && typeof mediaUrl === 'object' && (mediaUrl as any)._type === 'String') {
      mediaUrl = (mediaUrl as any).value
    }

    // Try previewUrl if uploaded file exists
    if (node.data.previewUrl) {
      let previewUrl = node.data.previewUrl
      if (previewUrl && typeof previewUrl === 'object' && (previewUrl as any)._type === 'String') {
        previewUrl = (previewUrl as any).value
      }
      mediaUrl = previewUrl
      console.log('Using previewUrl:', mediaUrl ? `${mediaUrl.substring(0, 50)}...` : 'null')
    }

    if (node.data.mediaId) {
      const mediaFile = getMediaFile(node.data.mediaId)
      if (mediaFile) {
        mediaUrl = mediaFile.dataUrl
        console.log('Using mediaFile.dataUrl:', mediaUrl ? `${mediaUrl.substring(0, 50)}...` : 'null')
      }
    }

    console.log('Final mediaUrl for video:', mediaUrl ? `${mediaUrl.substring(0, 50)}...` : 'null')

    if (mediaUrl) {
      const botMessage: ChatMessage = {
        id: `msg_${Date.now()}_${Math.random().toString(36).substring(2)}`,
        type: 'bot',
        content: node.data.message || '',
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

  private async handlePromptNode(node: FlowNode): Promise<void> {
    console.log('Processing AI prompt node:', node.data)
    
    const instance = node.data.instance || 'default'
    const openRouterKey = node.data.openRouterKey || ''

    // Analyze flow mode based on available data
    const flowMode = this.analyzeFlowMode(node)
    console.log(`Flow mode detected: ${flowMode}`)

    if (flowMode === 'AUTO') {
      // Full AI response mode - requires external AI service
      console.warn('AI functionality removed - falling back to manual response')
      await this.handleManualResponse(node, instance)
    } else {
      // Manual mode - use instance only
      await this.handleManualResponse(node, instance)
    }
  }

  private analyzeFlowMode(node: FlowNode): 'AUTO' | 'SEMI-AUTO' | 'MANUAL' {
    const openRouterKey = node.data.openRouterKey
    const instance = node.data.instance

    if (openRouterKey && instance) {
      return 'AUTO'
    } else if (instance && !openRouterKey) {
      return 'MANUAL'
    } else {
      return 'SEMI-AUTO'
    }
  }

  private async handleManualResponse(node: FlowNode, instance: string): Promise<void> {
    if (!this.execution.isWaitingForInput) {
      // First time reaching prompt node, wait for user input
      this.execution.isWaitingForInput = true
      this.onWaitingForInput()
      return
    }

    // Create manual response message
    const manualMessage: ChatMessage = {
      id: `msg_${Date.now()}_${Math.random().toString(36).substring(2)}`,
      type: 'bot',
      content: `Manual response from instance: ${instance}. (AI functionality has been removed)`,
      timestamp: new Date().toISOString()
    }

    this.execution.messages.push(manualMessage)
    this.onMessage(manualMessage)

    // Update execution variables
    this.execution.variables.instance = instance
    console.log('Manual response processed')
    await this.moveToNextNode()
  }

  private async moveToNextNode(): Promise<void> {
    const nextNode = this.getNextNode()
    console.log('Moving to next node:', nextNode)
    
    if (nextNode) {
      this.execution.currentNodeId = nextNode.id
      await this.processCurrentNode()
    } else {
      console.log('No next node found, completing execution')
      await this.completeExecution()
    }
  }

  private async saveState(): Promise<void> {
    // Save execution state to localStorage since MySQL bridge is removed
    try {
      await saveExecution(this.execution)
    } catch (error) {
      console.error('Error saving execution state:', error)
    }
  }

  private async completeExecution(): Promise<void> {
    this.execution.isCompleted = true
    this.execution.isWaitingForInput = false
    
    console.log('Flow execution completed')
    await this.saveState()
    this.onComplete()
  }
}