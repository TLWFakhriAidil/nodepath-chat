import "https://deno.land/x/xhr@0.1.0/mod.ts";
import { serve } from "https://deno.land/std@0.168.0/http/server.ts";

const corsHeaders = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type',
};

serve(async (req) => {
  // Handle CORS preflight requests
  if (req.method === 'OPTIONS') {
    return new Response(null, { headers: corsHeaders });
  }

  try {
    const OPENAI_API_KEY = Deno.env.get('OPENAI_API_KEY');
    
    if (!OPENAI_API_KEY) {
      throw new Error('OPENAI_API_KEY is not configured');
    }

    const { systemPrompt, userMessage, instance, openRouterKey, conversationHistory } = await req.json();

    console.log('AI Chat Request:', { 
      systemPrompt: systemPrompt?.substring(0, 100) + '...', 
      userMessage, 
      instance,
      openRouterKey: openRouterKey ? 'provided' : 'not provided',
      historyLength: conversationHistory?.length || 0
    });

    // Build conversation messages with history
    const messages = [
      { role: 'system', content: systemPrompt || 'You are a helpful assistant.' }
    ];

    // Add conversation history if available
    if (conversationHistory && Array.isArray(conversationHistory)) {
      conversationHistory.forEach((msg: any) => {
        if (msg.role === 'USER') {
          messages.push({ role: 'user', content: msg.content });
        } else if (msg.role === 'BOT') {
          messages.push({ role: 'assistant', content: msg.content });
        }
      });
    }

    // Add current user message
    messages.push({ role: 'user', content: userMessage });

    // Use OpenRouter if key is provided, otherwise use OpenAI
    const apiUrl = openRouterKey ? 'https://openrouter.ai/api/v1/chat/completions' : 'https://api.openai.com/v1/chat/completions';
    const apiKey = openRouterKey || OPENAI_API_KEY;
    const headers: any = {
      'Authorization': `Bearer ${apiKey}`,
      'Content-Type': 'application/json',
    };

    // Add OpenRouter specific headers
    if (openRouterKey) {
      headers['HTTP-Referer'] = 'https://lovableproject.com';
      headers['X-Title'] = 'Chatbot Test Environment';
    }

    const response = await fetch(apiUrl, {
      method: 'POST',
      headers,
      body: JSON.stringify({
        model: openRouterKey ? 'openai/gpt-4.1' : 'gpt-4.1-2025-04-14',
        messages,
        temperature: 0.7,
        max_tokens: 1000
      }),
    });

    if (!response.ok) {
      const errorData = await response.text();
      console.error('AI API Error:', errorData);
      throw new Error(`AI API error: ${response.status}`);
    }

    const data = await response.json();
    const aiReply = data.choices[0].message.content;

    console.log('AI Reply generated successfully');

    return new Response(JSON.stringify({ 
      success: true,
      aiReply,
      usage: data.usage
    }), {
      headers: { ...corsHeaders, 'Content-Type': 'application/json' },
    });

  } catch (error) {
    console.error('Error in test-ai-chat function:', error);
    return new Response(JSON.stringify({ 
      success: false,
      error: error.message 
    }), {
      status: 500,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' },
    });
  }
});