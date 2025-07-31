-- Create chatbot_flows table for storing chatbot flows
CREATE TABLE public.chatbot_flows (
  id TEXT NOT NULL PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT DEFAULT '',
  nodes JSONB NOT NULL DEFAULT '[]',
  edges JSONB NOT NULL DEFAULT '[]',
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Enable Row Level Security
ALTER TABLE public.chatbot_flows ENABLE ROW LEVEL SECURITY;

-- Create policies for public access (matching current localStorage behavior)
CREATE POLICY "Allow all operations for everyone" 
ON public.chatbot_flows 
FOR ALL 
USING (true);

-- Create trigger for automatic timestamp updates
CREATE TRIGGER update_chatbot_flows_updated_at
BEFORE UPDATE ON public.chatbot_flows
FOR EACH ROW
EXECUTE FUNCTION public.update_updated_at_column();

-- Create chatbot_executions table for storing flow executions
CREATE TABLE public.chatbot_executions (
  id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
  flow_id TEXT NOT NULL,
  current_node_id TEXT NOT NULL,
  variables JSONB NOT NULL DEFAULT '{}',
  messages JSONB NOT NULL DEFAULT '[]',
  is_waiting_for_input BOOLEAN NOT NULL DEFAULT false,
  is_completed BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  UNIQUE(flow_id)
);

-- Enable Row Level Security
ALTER TABLE public.chatbot_executions ENABLE ROW LEVEL SECURITY;

-- Create policies for public access
CREATE POLICY "Allow all operations for everyone" 
ON public.chatbot_executions 
FOR ALL 
USING (true);

-- Create trigger for automatic timestamp updates
CREATE TRIGGER update_chatbot_executions_updated_at
BEFORE UPDATE ON public.chatbot_executions
FOR EACH ROW
EXECUTE FUNCTION public.update_updated_at_column();