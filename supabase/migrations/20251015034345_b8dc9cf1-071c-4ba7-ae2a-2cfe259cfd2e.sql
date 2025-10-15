-- Create custom types
CREATE TYPE public.api_provider AS ENUM ('openai', 'anthropic', 'google', 'openrouter');
CREATE TYPE public.whatsapp_provider AS ENUM ('waha', 'whacenter', 'wablas', 'wasapbot');
CREATE TYPE public.device_status AS ENUM ('active', 'inactive', 'pending');
CREATE TYPE public.flow_status AS ENUM ('active', 'inactive', 'draft');
CREATE TYPE public.message_type AS ENUM ('text', 'image', 'audio', 'video', 'document');
CREATE TYPE public.human_status AS ENUM ('active', 'inactive', 'pending_response');

-- Create profiles table for user data
CREATE TABLE public.profiles (
  id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  email TEXT,
  full_name TEXT,
  avatar_url TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable RLS on profiles
ALTER TABLE public.profiles ENABLE ROW LEVEL SECURITY;

-- Profiles policies
CREATE POLICY "Users can view their own profile"
  ON public.profiles FOR SELECT
  USING (auth.uid() = id);

CREATE POLICY "Users can update their own profile"
  ON public.profiles FOR UPDATE
  USING (auth.uid() = id);

-- Create device_settings table
CREATE TABLE public.device_settings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE NOT NULL,
  device_id TEXT NOT NULL,
  device_name TEXT NOT NULL,
  api_key_option TEXT,
  api_provider public.api_provider DEFAULT 'openrouter',
  whatsapp_provider public.whatsapp_provider DEFAULT 'waha',
  waha_url TEXT,
  whacenter_token TEXT,
  status public.device_status DEFAULT 'active',
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  UNIQUE(user_id, device_id)
);

ALTER TABLE public.device_settings ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view their own devices"
  ON public.device_settings FOR SELECT
  USING (auth.uid() = user_id);

CREATE POLICY "Users can create their own devices"
  ON public.device_settings FOR INSERT
  WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update their own devices"
  ON public.device_settings FOR UPDATE
  USING (auth.uid() = user_id);

CREATE POLICY "Users can delete their own devices"
  ON public.device_settings FOR DELETE
  USING (auth.uid() = user_id);

-- Create chatbot_flows table
CREATE TABLE public.chatbot_flows (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE NOT NULL,
  device_id UUID REFERENCES public.device_settings(id) ON DELETE SET NULL,
  flow_id TEXT NOT NULL,
  flow_name TEXT NOT NULL,
  description TEXT,
  niche TEXT,
  nodes JSONB DEFAULT '[]'::jsonb,
  edges JSONB DEFAULT '[]'::jsonb,
  global_instance TEXT,
  global_open_router_key TEXT,
  status public.flow_status DEFAULT 'active',
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  UNIQUE(user_id, flow_id)
);

ALTER TABLE public.chatbot_flows ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view their own flows"
  ON public.chatbot_flows FOR SELECT
  USING (auth.uid() = user_id);

CREATE POLICY "Users can create their own flows"
  ON public.chatbot_flows FOR INSERT
  WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update their own flows"
  ON public.chatbot_flows FOR UPDATE
  USING (auth.uid() = user_id);

CREATE POLICY "Users can delete their own flows"
  ON public.chatbot_flows FOR DELETE
  USING (auth.uid() = user_id);

-- Create ai_conversations table
CREATE TABLE public.ai_conversations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE NOT NULL,
  device_id UUID REFERENCES public.device_settings(id) ON DELETE CASCADE,
  flow_id UUID REFERENCES public.chatbot_flows(id) ON DELETE SET NULL,
  prospect_num TEXT NOT NULL,
  prospect_name TEXT,
  conversation_current TEXT,
  conversation_last TEXT,
  stage TEXT,
  current_node_id TEXT,
  waiting_for_reply BOOLEAN DEFAULT FALSE,
  human_status public.human_status DEFAULT 'active',
  session_locked BOOLEAN DEFAULT FALSE,
  session_locked_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  UNIQUE(device_id, prospect_num)
);

ALTER TABLE public.ai_conversations ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view their own conversations"
  ON public.ai_conversations FOR SELECT
  USING (auth.uid() = user_id);

CREATE POLICY "Users can create their own conversations"
  ON public.ai_conversations FOR INSERT
  WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update their own conversations"
  ON public.ai_conversations FOR UPDATE
  USING (auth.uid() = user_id);

CREATE POLICY "Users can delete their own conversations"
  ON public.ai_conversations FOR DELETE
  USING (auth.uid() = user_id);

-- Create conversation_logs table
CREATE TABLE public.conversation_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE NOT NULL,
  device_id UUID REFERENCES public.device_settings(id) ON DELETE CASCADE,
  phone_number TEXT NOT NULL,
  message_type public.message_type DEFAULT 'text',
  content TEXT,
  ai_response_time INTERVAL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

ALTER TABLE public.conversation_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view their own logs"
  ON public.conversation_logs FOR SELECT
  USING (auth.uid() = user_id);

CREATE POLICY "Users can create their own logs"
  ON public.conversation_logs FOR INSERT
  WITH CHECK (auth.uid() = user_id);

-- Create analytics_summary table
CREATE TABLE public.analytics_summary (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE NOT NULL,
  device_id UUID REFERENCES public.device_settings(id) ON DELETE CASCADE,
  date_period DATE NOT NULL,
  total_conversations INTEGER DEFAULT 0,
  total_messages INTEGER DEFAULT 0,
  avg_response_time INTERVAL,
  active_users INTEGER DEFAULT 0,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  UNIQUE(device_id, date_period)
);

ALTER TABLE public.analytics_summary ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view their own analytics"
  ON public.analytics_summary FOR SELECT
  USING (auth.uid() = user_id);

-- Create indexes for performance
CREATE INDEX idx_device_settings_user_id ON public.device_settings(user_id);
CREATE INDEX idx_device_settings_device_id ON public.device_settings(device_id);
CREATE INDEX idx_device_settings_status ON public.device_settings(status);

CREATE INDEX idx_chatbot_flows_user_id ON public.chatbot_flows(user_id);
CREATE INDEX idx_chatbot_flows_flow_id ON public.chatbot_flows(flow_id);
CREATE INDEX idx_chatbot_flows_device_id ON public.chatbot_flows(device_id);
CREATE INDEX idx_chatbot_flows_status ON public.chatbot_flows(status);

CREATE INDEX idx_ai_conversations_user_id ON public.ai_conversations(user_id);
CREATE INDEX idx_ai_conversations_device_id ON public.ai_conversations(device_id);
CREATE INDEX idx_ai_conversations_prospect_num ON public.ai_conversations(prospect_num);
CREATE INDEX idx_ai_conversations_flow_id ON public.ai_conversations(flow_id);
CREATE INDEX idx_ai_conversations_stage ON public.ai_conversations(stage);
CREATE INDEX idx_ai_conversations_waiting_for_reply ON public.ai_conversations(waiting_for_reply);

CREATE INDEX idx_conversation_logs_user_id ON public.conversation_logs(user_id);
CREATE INDEX idx_conversation_logs_device_id ON public.conversation_logs(device_id);
CREATE INDEX idx_conversation_logs_phone_number ON public.conversation_logs(phone_number);
CREATE INDEX idx_conversation_logs_created_at ON public.conversation_logs(created_at DESC);

CREATE INDEX idx_analytics_summary_user_id ON public.analytics_summary(user_id);
CREATE INDEX idx_analytics_summary_device_id ON public.analytics_summary(device_id);
CREATE INDEX idx_analytics_summary_date_period ON public.analytics_summary(date_period DESC);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION public.handle_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add updated_at triggers
CREATE TRIGGER set_updated_at_profiles
  BEFORE UPDATE ON public.profiles
  FOR EACH ROW EXECUTE FUNCTION public.handle_updated_at();

CREATE TRIGGER set_updated_at_device_settings
  BEFORE UPDATE ON public.device_settings
  FOR EACH ROW EXECUTE FUNCTION public.handle_updated_at();

CREATE TRIGGER set_updated_at_chatbot_flows
  BEFORE UPDATE ON public.chatbot_flows
  FOR EACH ROW EXECUTE FUNCTION public.handle_updated_at();

CREATE TRIGGER set_updated_at_ai_conversations
  BEFORE UPDATE ON public.ai_conversations
  FOR EACH ROW EXECUTE FUNCTION public.handle_updated_at();

CREATE TRIGGER set_updated_at_analytics_summary
  BEFORE UPDATE ON public.analytics_summary
  FOR EACH ROW EXECUTE FUNCTION public.handle_updated_at();

-- Create profile auto-creation trigger
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO public.profiles (id, email, full_name)
  VALUES (
    NEW.id,
    NEW.email,
    COALESCE(NEW.raw_user_meta_data->>'full_name', '')
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();