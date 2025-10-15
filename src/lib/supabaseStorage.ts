import { supabase } from '@/integrations/supabase/client';
import { ChatbotFlow } from '@/types/chatbot';

// Flow management using Supabase
export const saveFlow = async (flow: ChatbotFlow): Promise<void> => {
  try {
    // Validate required parameters
    if (!flow.id || !flow.name) {
      throw new Error('Flow ID and name are required for saving');
    }

    if (!flow.nodes || !Array.isArray(flow.nodes) || flow.nodes.length === 0) {
      throw new Error('Flow must have at least one node');
    }

    console.log('Saving flow to Supabase:', flow.id);
    
    // Get current user
    const { data: { user } } = await supabase.auth.getUser();
    if (!user) {
      throw new Error('User must be authenticated to save flows');
    }

    // Check if flow exists
    const { data: existingFlow } = await supabase
      .from('chatbot_flows')
      .select('id')
      .eq('id', flow.id)
      .maybeSingle();

    const flowData = {
      id: flow.id,
      flow_id: flow.id,
      flow_name: flow.name,
      description: flow.description || '',
      niche: flow.niche || '',
      device_id: flow.selectedDeviceId || flow.id_device || flow.idDevice || null,
      nodes: flow.nodes as any,
      edges: flow.edges as any || [],
      user_id: user.id,
      status: 'active' as const,
      global_instance: null,
      global_open_router_key: null
    };

    if (existingFlow) {
      // Update existing flow
      const { error } = await supabase
        .from('chatbot_flows')
        .update(flowData)
        .eq('id', flow.id);

      if (error) throw error;
    } else {
      // Insert new flow
      const { error } = await supabase
        .from('chatbot_flows')
        .insert(flowData);

      if (error) throw error;
    }

    console.log('Flow saved successfully to Supabase:', flow.id);
  } catch (error) {
    console.error('Error saving flow to Supabase:', error);
    throw error;
  }
};

export const getFlows = async (): Promise<ChatbotFlow[]> => {
  try {
    const { data: { user } } = await supabase.auth.getUser();
    if (!user) {
      throw new Error('User must be authenticated');
    }

    const { data, error } = await supabase
      .from('chatbot_flows')
      .select('*')
      .eq('user_id', user.id)
      .order('created_at', { ascending: false });

    if (error) throw error;

    return (data || []).map((row: any) => ({
      id: row.id,
      name: row.flow_name,
      description: row.description || '',
      niche: row.niche || '',
      selectedDeviceId: row.device_id || '',
      nodes: (row.nodes as any) || [],
      edges: (row.edges as any) || [],
      createdAt: row.created_at,
      updatedAt: row.updated_at
    }));
  } catch (error) {
    console.error('Error fetching flows from Supabase:', error);
    return [];
  }
};

export const getFlow = async (id: string): Promise<ChatbotFlow | null> => {
  try {
    const { data: { user } } = await supabase.auth.getUser();
    if (!user) {
      throw new Error('User must be authenticated');
    }

    const { data, error } = await supabase
      .from('chatbot_flows')
      .select('*')
      .eq('id', id)
      .eq('user_id', user.id)
      .maybeSingle();

    if (error) throw error;
    if (!data) return null;

    return {
      id: data.id,
      name: data.flow_name,
      description: data.description || '',
      niche: data.niche || '',
      selectedDeviceId: data.device_id || '',
      nodes: (data.nodes as any) || [],
      edges: (data.edges as any) || [],
      createdAt: data.created_at,
      updatedAt: data.updated_at
    };
  } catch (error) {
    console.error('Error fetching flow from Supabase:', error);
    return null;
  }
};

export const deleteFlow = async (id: string): Promise<void> => {
  try {
    const { data: { user } } = await supabase.auth.getUser();
    if (!user) {
      throw new Error('User must be authenticated');
    }

    const { error } = await supabase
      .from('chatbot_flows')
      .delete()
      .eq('id', id)
      .eq('user_id', user.id);

    if (error) throw error;

    console.log('Flow deleted successfully from Supabase');
  } catch (error) {
    console.error('Error deleting flow from Supabase:', error);
    throw error;
  }
};

// Media files still use localStorage for now
export const saveMediaFile = async (file: any): Promise<any> => {
  try {
    const mediaFiles = JSON.parse(localStorage.getItem('media_files') || '[]');
    mediaFiles.push({
      ...file,
      uploaded_at: file.uploaded_at || new Date().toISOString(),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    });
    localStorage.setItem('media_files', JSON.stringify(mediaFiles));
    return file;
  } catch (error) {
    console.error('Error saving media file:', error);
    throw error;
  }
};

export const getMediaFiles = async (): Promise<any[]> => {
  try {
    const mediaFiles = JSON.parse(localStorage.getItem('media_files') || '[]');
    return mediaFiles.sort((a: any, b: any) => 
      new Date(b.uploaded_at).getTime() - new Date(a.uploaded_at).getTime()
    );
  } catch (error) {
    console.error('Error fetching media files:', error);
    return [];
  }
};

export const getMediaFile = async (id: string): Promise<any | null> => {
  try {
    const mediaFiles = JSON.parse(localStorage.getItem('media_files') || '[]');
    return mediaFiles.find((f: any) => f.id === id) || null;
  } catch (error) {
    console.error('Error fetching media file:', error);
    return null;
  }
};

export const deleteMediaFile = async (id: string): Promise<void> => {
  try {
    const mediaFiles = JSON.parse(localStorage.getItem('media_files') || '[]');
    const filteredFiles = mediaFiles.filter((f: any) => f.id !== id);
    localStorage.setItem('media_files', JSON.stringify(filteredFiles));
  } catch (error) {
    console.error('Error deleting media file:', error);
    throw error;
  }
};

// Flow execution still uses localStorage
export const saveFlowExecution = async (execution: any): Promise<void> => {
  try {
    const simulationId = execution.id || `exec_${execution.flowId}_${Date.now()}_${Math.random().toString(36).substring(2)}`;
    
    const executions = JSON.parse(localStorage.getItem('flow_executions') || '{}');
    
    executions[simulationId] = {
      ...execution,
      id: simulationId,
      updated_at: new Date().toISOString()
    };
    
    localStorage.setItem('flow_executions', JSON.stringify(executions));
    console.log('Flow execution saved to localStorage:', simulationId);
  } catch (error) {
    console.error('Error saving flow execution:', error);
    throw error;
  }
};

export const getFlowExecution = async (id: string): Promise<any | null> => {
  try {
    const executions = JSON.parse(localStorage.getItem('flow_executions') || '{}');
    return executions[id] || null;
  } catch (error) {
    console.error('Error fetching flow execution:', error);
    return null;
  }
};

export const updateFlowExecution = async (id: string, updates: any): Promise<void> => {
  try {
    const executions = JSON.parse(localStorage.getItem('flow_executions') || '{}');
    
    if (executions[id]) {
      executions[id] = {
        ...executions[id],
        ...updates,
        updated_at: new Date().toISOString()
      };
      localStorage.setItem('flow_executions', JSON.stringify(executions));
    }
  } catch (error) {
    console.error('Error updating flow execution:', error);
    throw error;
  }
};

export const deleteFlowExecution = async (id: string): Promise<void> => {
  try {
    const executions = JSON.parse(localStorage.getItem('flow_executions') || '{}');
    delete executions[id];
    localStorage.setItem('flow_executions', JSON.stringify(executions));
  } catch (error) {
    console.error('Error deleting flow execution:', error);
    throw error;
  }
};

// Helper function to extract AI prompt data from flow nodes
export const extractAIPromptData = (flow: ChatbotFlow) => {
  const aiNodes = flow.nodes.filter(node => node.type === 'prompt');
  return aiNodes.map(node => ({
    nodeId: node.id,
    instance: node.data.instance || '',
    openRouterKey: node.data.openRouterKey || ''
  }));
};
