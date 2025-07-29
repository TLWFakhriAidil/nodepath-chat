import React, { useState, useRef, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Send, Bot, User, Play, Square } from 'lucide-react'
import { ChatMessage, FlowExecution } from '@/types/chatbot'
import { FlowEngine } from '@/lib/flowEngine'
import { getFlows } from '@/lib/localStorage'
import { useToast } from '@/hooks/use-toast'

const ChatSimulation = ({ preselectedFlowId }: { preselectedFlowId?: string | null }) => {
  const [selectedFlowId, setSelectedFlowId] = useState<string>(preselectedFlowId || '')
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [currentInput, setCurrentInput] = useState('')
  const [isRunning, setIsRunning] = useState(false)
  const [isWaitingForInput, setIsWaitingForInput] = useState(false)
  const [flowEngine, setFlowEngine] = useState<FlowEngine | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const { toast } = useToast()

  const flows = getFlows()
  console.log('Available flows in simulation:', flows)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  useEffect(() => {
    if (preselectedFlowId) {
      setSelectedFlowId(preselectedFlowId)
    }
  }, [preselectedFlowId])

  const startFlow = async () => {
    if (!selectedFlowId) {
      toast({
        title: "No flow selected",
        description: "Please select a flow to test",
        variant: "destructive"
      })
      return
    }

    try {
      const engine = new FlowEngine(
        selectedFlowId,
        (message) => {
          setMessages(prev => [...prev, message])
        },
        () => {
          setIsRunning(false)
          setIsWaitingForInput(false)
          toast({
            title: "Flow completed",
            description: "The chatbot flow has finished executing"
          })
        },
        () => {
          setIsWaitingForInput(true)
        }
      )

      setFlowEngine(engine)
      setMessages([])
      setIsRunning(true)
      setIsWaitingForInput(false)

      await engine.start()
    } catch (error) {
      toast({
        title: "Error starting flow",
        description: error instanceof Error ? error.message : 'Unknown error',
        variant: "destructive"
      })
    }
  }

  const stopFlow = () => {
    setFlowEngine(null)
    setIsRunning(false)
    setIsWaitingForInput(false)
    setMessages([])
  }

  const sendMessage = async () => {
    if (!currentInput.trim() || !flowEngine || !isWaitingForInput) {
      return
    }

    const input = currentInput.trim()
    setCurrentInput('')

    try {
      await flowEngine.processUserInput(input)
      setIsWaitingForInput(flowEngine.isWaitingForInput())
    } catch (error) {
      toast({
        title: "Error processing input",
        description: error instanceof Error ? error.message : 'Unknown error',
        variant: "destructive"
      })
    }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }

  const renderMessage = (message: ChatMessage) => {
    const isBot = message.type === 'bot'
    
    return (
      <div key={message.id} className={`flex gap-3 ${isBot ? 'justify-start' : 'justify-end'}`}>
        {isBot && (
          <Avatar className="w-8 h-8">
            <AvatarFallback className="bg-primary text-primary-foreground">
              <Bot className="w-4 h-4" />
            </AvatarFallback>
          </Avatar>
        )}
        
        <div className={`max-w-[70%] ${isBot ? 'order-2' : 'order-1'}`}>
          <div className={`rounded-lg p-3 ${
            isBot 
              ? 'bg-muted text-muted-foreground' 
              : 'bg-primary text-primary-foreground ml-auto'
          }`}>
            {message.content && (
              <p className="text-sm whitespace-pre-wrap">{message.content}</p>
            )}
            
            {message.mediaType === 'image' && message.mediaUrl && (
              <div className="mt-2">
                <img 
                  src={message.mediaUrl} 
                  alt="Shared image" 
                  className="max-w-full h-auto rounded-md"
                />
              </div>
            )}
            
            {message.mediaType === 'audio' && message.mediaUrl && (
              <div className="mt-2">
                <audio controls className="w-full">
                  <source src={message.mediaUrl} />
                  Your browser does not support the audio element.
                </audio>
              </div>
            )}
            
            {message.mediaType === 'video' && message.mediaUrl && (
              <div className="mt-2">
                <video controls className="w-full max-w-sm rounded-md">
                  <source src={message.mediaUrl} />
                  Your browser does not support the video element.
                </video>
              </div>
            )}
          </div>
          
          <div className="text-xs text-muted-foreground mt-1 px-1">
            {new Date(message.timestamp).toLocaleTimeString()}
          </div>
        </div>
        
        {!isBot && (
          <Avatar className="w-8 h-8 order-3">
            <AvatarFallback className="bg-secondary text-secondary-foreground">
              <User className="w-4 h-4" />
            </AvatarFallback>
          </Avatar>
        )}
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col max-w-4xl mx-auto p-6">
      <Card className="flex-1 flex flex-col">
        <CardHeader className="flex-shrink-0">
          <CardTitle className="flex items-center justify-between">
            <span>Chat Simulation</span>
            <div className="flex items-center gap-2">
              {isRunning && (
                <Badge variant={isWaitingForInput ? "secondary" : "default"}>
                  {isWaitingForInput ? "Waiting for input" : "Running"}
                </Badge>
              )}
            </div>
          </CardTitle>
          
          <div className="flex items-center gap-4">
            <Select value={selectedFlowId} onValueChange={setSelectedFlowId} disabled={isRunning}>
              <SelectTrigger className="flex-1">
                <SelectValue placeholder="Select a flow to test" />
              </SelectTrigger>
              <SelectContent>
                {flows.map((flow) => (
                  <SelectItem key={flow.id} value={flow.id}>
                    {flow.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            
            {!isRunning ? (
              <Button onClick={startFlow} disabled={!selectedFlowId}>
                <Play className="w-4 h-4 mr-2" />
                Start Flow
              </Button>
            ) : (
              <Button onClick={stopFlow} variant="destructive">
                <Square className="w-4 h-4 mr-2" />
                Stop
              </Button>
            )}
          </div>
        </CardHeader>
        
        <CardContent className="flex-1 flex flex-col p-0">
          {/* Messages Area */}
          <div className="flex-1 overflow-y-auto p-6 space-y-4 min-h-0">
            {messages.length === 0 ? (
              <div className="flex items-center justify-center h-full text-muted-foreground">
                {isRunning 
                  ? "Flow is starting..." 
                  : flows.length === 0
                    ? "No flows available. Create a flow in the Flow Builder first."
                    : "Select a flow and click Start Flow to begin testing"
                }
              </div>
            ) : (
              <>
                {messages.map(renderMessage)}
                <div ref={messagesEndRef} />
              </>
            )}
          </div>
          
          {/* Input Area */}
          {isRunning && (
            <div className="border-t p-4">
              <div className="flex gap-2">
                <Input
                  value={currentInput}
                  onChange={(e) => setCurrentInput(e.target.value)}
                  onKeyPress={handleKeyPress}
                  placeholder={isWaitingForInput ? "Type your response..." : "Waiting for bot..."}
                  disabled={!isWaitingForInput}
                  className="flex-1"
                />
                <Button 
                  onClick={sendMessage} 
                  disabled={!isWaitingForInput || !currentInput.trim()}
                  size="icon"
                >
                  <Send className="w-4 h-4" />
                </Button>
              </div>
              
              {!isWaitingForInput && (
                <p className="text-xs text-muted-foreground mt-2">
                  The bot is processing. Please wait...
                </p>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

export default ChatSimulation