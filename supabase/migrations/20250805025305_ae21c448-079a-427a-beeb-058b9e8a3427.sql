-- Create chatbot_executions_nodepath table for Test Chat feature
CREATE TABLE IF NOT EXISTS public.chatbot_executions_nodepath (
  id VARCHAR(255) PRIMARY KEY,
  system_prompt TEXT,
  instance VARCHAR(255),
  open_router_key VARCHAR(255),
  conv_last JSONB,
  conv_current TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Enable Row Level Security
ALTER TABLE public.chatbot_executions_nodepath ENABLE ROW LEVEL SECURITY;

-- Create policy to allow all operations for everyone (same as other tables)
CREATE POLICY "Allow all operations for everyone" 
ON public.chatbot_executions_nodepath 
FOR ALL 
USING (true);

-- Create trigger for automatic timestamp updates
CREATE TRIGGER update_chatbot_executions_nodepath_updated_at
BEFORE UPDATE ON public.chatbot_executions_nodepath
FOR EACH ROW
EXECUTE FUNCTION public.update_updated_at_column();