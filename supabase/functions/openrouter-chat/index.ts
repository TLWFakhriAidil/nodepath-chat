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
    const { systemPrompt, userMessage, conversationHistory, openRouterKey, instance } = await req.json();

    console.log('OpenRouter Chat request:', {
      systemPrompt: systemPrompt?.substring(0, 100) + '...',
      userMessage,
      historyLength: conversationHistory?.length || 0,
      instance,
      hasKey: !!openRouterKey
    });

    if (!openRouterKey) {
      throw new Error('OpenRouter API key is required');
    }

    if (!systemPrompt && !userMessage) {
      throw new Error('Either systemPrompt or userMessage is required');
    }

    // Prepare messages for OpenAI format
    const messages = [];
    
    // Add system prompt if provided
    if (systemPrompt) {
      messages.push({
        role: 'system',
        content: systemPrompt
      });
    }

    // Add conversation history if provided
    if (conversationHistory && Array.isArray(conversationHistory)) {
      conversationHistory.forEach((msg: any) => {
        if (msg.role && msg.content) {
          messages.push({
            role: msg.role === 'bot' ? 'assistant' : msg.role,
            content: msg.content
          });
        }
      });
    }

    // Add current user message
    if (userMessage) {
      messages.push({
        role: 'user',
        content: userMessage
      });
    }

    console.log('Sending to OpenRouter:', {
      messageCount: messages.length,
      model: 'openai/gpt-4.1-2025-04-14'
    });

    // Call OpenRouter API
    const response = await fetch('https://openrouter.ai/api/v1/chat/completions', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${openRouterKey}`,
        'Content-Type': 'application/json',
        'HTTP-Referer': 'https://your-app.com',
        'X-Title': 'AI Chatbot Flow'
      },
      body: JSON.stringify({
        model: 'openai/gpt-4.1-2025-04-14',
        messages: messages,
        temperature: 0.7,
        max_tokens: 1000,
        top_p: 0.9
      }),
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error('OpenRouter API error:', errorText);
      throw new Error(`OpenRouter API error: ${response.status} - ${errorText}`);
    }

    const data = await response.json();
    console.log('OpenRouter response received');

    const aiReply = data.choices?.[0]?.message?.content || 'Sorry, I could not generate a response.';

    // Prepare conversation data for storage
    const conversationData = {
      conv_current: aiReply,
      conv_last: [...(conversationHistory || []), 
        { role: 'user', content: userMessage },
        { role: 'assistant', content: aiReply }
      ],
      instance: instance || 'default',
      usage: data.usage || {}
    };

    return new Response(JSON.stringify({
      success: true,
      reply: aiReply,
      conversationData,
      usage: data.usage
    }), {
      headers: { ...corsHeaders, 'Content-Type': 'application/json' },
    });

  } catch (error) {
    console.error('Error in openrouter-chat function:', error);
    return new Response(JSON.stringify({ 
      success: false,
      error: error.message,
      reply: 'Sorry, I encountered an error while processing your request.'
    }), {
      status: 500,
      headers: { ...corsHeaders, 'Content-Type': 'application/json' },
    });
  }
});