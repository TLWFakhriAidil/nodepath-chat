-- Create chatbot_executions_nodepath table for Test Chat feature
CREATE TABLE IF NOT EXISTS public.chatbot_executions_nodepath (
  id VARCHAR(255) PRIMARY KEY,
  system_prompt TEXT,
  instance VARCHAR(255),
  open_router_key VARCHAR(255),
  conv_last JSON,
  conv_current TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Enable Row Level Security
ALTER TABLE public.chatbot_executions_nodepath ENABLE ROW LEVEL SECURITY;

-- Create policy to allow all operations for everyone (same as other tables)
CREATE POLICY "Allow all operations for everyone" 
ON public.chatbot_executions_nodepath 
FOR ALL 
USING (true);