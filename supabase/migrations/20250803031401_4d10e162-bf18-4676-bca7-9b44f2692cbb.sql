-- Create AI settings table for chatbot configuration
CREATE TABLE IF NOT EXISTS public.ai_settings_nodepath (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    id_staff TEXT,
    system_prompt TEXT,
    closing_prompt TEXT,
    instance_prompt TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Enable RLS
ALTER TABLE public.ai_settings_nodepath ENABLE ROW LEVEL SECURITY;

-- Create policies for public access (matching existing pattern)
CREATE POLICY "Allow all operations for everyone" 
ON public.ai_settings_nodepath 
FOR ALL 
USING (true);

-- Create trigger for automatic timestamp updates
CREATE TRIGGER update_ai_settings_updated_at
BEFORE UPDATE ON public.ai_settings_nodepath
FOR EACH ROW
EXECUTE FUNCTION public.update_updated_at_column();