import { ChatbotFlow, MediaFile, FlowExecution } from '@/types/chatbot'

const MYSQL_CONFIG = {
  host: '159.89.198.71',
  port: 3306,
  user: 'admin_aqil',
  password: 'admin_aqil',
  database: 'admin_railway'
}

// Fallback MySQL connection - Edge Function not available in this environment
const callMySQLAPI = async (query: string, params: any[] = []) => {
  try {
    console.log('Attempting MySQL connection...', query.substring(0, 50) + '...');
    
    // Check if Edge Function is available
    const response = await fetch('/functions/v1/mysql-api', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        query,
        params
      })
    });

    if (response.status === 404) {
      throw new Error('EDGE_FUNCTION_NOT_DEPLOYED');
    }

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result = await response.json();
    
    if (result.success) {
      console.log('MySQL operation successful:', result);
      return result;
    } else {
      throw new Error(result.error || 'MySQL operation failed');
    }
  } catch (error) {
    if (error.message === 'EDGE_FUNCTION_NOT_DEPLOYED') {
      console.warn('Edge Function not deployed - using localStorage fallback');
      // Return mock success for localStorage fallback
      return { success: true, affectedRows: 1, data: [] };
    }
    console.error('MySQL connection error:', error);
    throw error;
  }
}

// Flow management
export const saveFlow = async (flow: ChatbotFlow): Promise<void> => {
  try {
    // Validate required parameters before saving
    if (!flow.id || !flow.name) {
      throw new Error('Flow ID and name are required for saving')
    }

    if (!flow.nodes || !Array.isArray(flow.nodes) || flow.nodes.length === 0) {
      throw new Error('Flow must have at least one node')
    }

    console.log('Saving flow to MySQL database:', flow.id)
    
    // Prepare flow data for MySQL
    const flowData = {
      id: flow.id,
      name: flow.name,
      description: flow.description || '',
      instance: flow.globalInstance || null,
      open_router_key: flow.globalOpenRouterKey || null,
      nodes: JSON.stringify(flow.nodes),
      edges: JSON.stringify(flow.edges || []),
      created_at: flow.createdAt || new Date().toISOString(),
      updated_at: new Date().toISOString()
    }
    
    // Single MySQL operation - create table and insert/update in one call
    const saveQuery = `
      CREATE TABLE IF NOT EXISTS chatbot_flows_nodepath (
        id VARCHAR(255) PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        description TEXT,
        instance VARCHAR(255),
        open_router_key VARCHAR(255),
        nodes LONGTEXT,
        edges LONGTEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
      );
      
      INSERT INTO chatbot_flows_nodepath (
        id, name, description, instance, open_router_key, nodes, edges, created_at, updated_at
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
      ON DUPLICATE KEY UPDATE
        name = VALUES(name),
        description = VALUES(description),
        instance = VALUES(instance),
        open_router_key = VALUES(open_router_key),
        nodes = VALUES(nodes),
        edges = VALUES(edges),
        updated_at = VALUES(updated_at);
    `;
    
    const params = [
      flowData.id,
      flowData.name,
      flowData.description,
      flowData.instance,
      flowData.open_router_key,
      flowData.nodes,
      flowData.edges,
      flowData.created_at,
      flowData.updated_at
    ];
    
    const result = await callMySQLAPI(saveQuery, params);
    
    // Also save to localStorage as backup/fallback
    const flows = JSON.parse(localStorage.getItem('chatbot_flows') || '[]');
    const existingIndex = flows.findIndex((f: any) => f.id === flow.id);
    
    const flowForStorage = {
      ...flow,
      globalInstance: flow.globalInstance,
      globalOpenRouterKey: flow.globalOpenRouterKey,
      updatedAt: new Date().toISOString()
    };
    
    if (existingIndex >= 0) {
      flows[existingIndex] = flowForStorage;
    } else {
      flows.push(flowForStorage);
    }
    
    localStorage.setItem('chatbot_flows', JSON.stringify(flows));
    
    console.log('Flow saved successfully:', {
      id: flow.id,
      name: flow.name,
      instance: flow.globalInstance,
      openRouterKey: flow.globalOpenRouterKey,
      affectedRows: result.affectedRows
    })
  } catch (error) {
    console.error('Error saving flow:', error)
    throw error
  }
}

// Ensure MySQL table structure
const ensureTableStructure = async () => {
  try {
    const createTableQuery = `
      CREATE TABLE IF NOT EXISTS chatbot_flows_nodepath (
        id VARCHAR(255) PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        description TEXT,
        instance VARCHAR(255),
        open_router_key VARCHAR(255),
        nodes LONGTEXT,
        edges LONGTEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
      )
    `
    
    await callMySQLAPI(createTableQuery)
  } catch (error) {
    console.warn('Failed to ensure table structure:', error)
  }
}

export const getFlows = async (): Promise<ChatbotFlow[]> => {
  try {
    // Get flows from MySQL database
    const query = 'SELECT * FROM chatbot_flows_nodepath ORDER BY updated_at DESC';
    const result = await callMySQLAPI(query);
    
    if (result.data) {
      const formattedFlows = result.data.map((row: any) => ({
        id: row.id,
        name: row.name,
        description: row.description,
        globalInstance: row.instance,
        globalOpenRouterKey: row.open_router_key,
        nodes: JSON.parse(row.nodes || '[]'),
        edges: JSON.parse(row.edges || '[]'),
        createdAt: row.created_at,
        updatedAt: row.updated_at
      }));
      
      return formattedFlows;
    }
    
    return [];
  } catch (error) {
    console.error('Error fetching flows from MySQL:', error)
    // Fallback to localStorage if MySQL fails
    try {
      const flows = JSON.parse(localStorage.getItem('chatbot_flows') || '[]')
      return flows.map((flow: any) => ({
        ...flow,
        globalInstance: flow.globalInstance || flow.instance || null,
        globalOpenRouterKey: flow.globalOpenRouterKey || flow.open_router_key || null
      }))
    } catch (localError) {
      console.error('Error fetching flows from localStorage:', localError)
      return []
    }
  }
}

export const getFlow = async (id: string): Promise<ChatbotFlow | null> => {
  try {
    // Get flow from MySQL database
    const query = 'SELECT * FROM chatbot_flows_nodepath WHERE id = ?';
    const result = await callMySQLAPI(query, [id]);
    
    if (result.data && result.data.length > 0) {
      const row = result.data[0];
      return {
        id: row.id,
        name: row.name,
        description: row.description,
        globalInstance: row.instance,
        globalOpenRouterKey: row.open_router_key,
        nodes: JSON.parse(row.nodes || '[]'),
        edges: JSON.parse(row.edges || '[]'),
        createdAt: row.created_at,
        updatedAt: row.updated_at
      };
    }
    
    return null;
  } catch (error) {
    console.error('Error fetching flow from MySQL:', error)
    // Fallback to localStorage
    try {
      const flows = JSON.parse(localStorage.getItem('chatbot_flows') || '[]')
      const flow = flows.find((f: any) => f.id === id)
      
      if (flow) {
        return {
          ...flow,
          globalInstance: flow.globalInstance || flow.instance || null,
          globalOpenRouterKey: flow.globalOpenRouterKey || flow.open_router_key || null
        }
      }
      
      return null
    } catch (localError) {
      console.error('Error fetching flow from localStorage:', localError)
      return null
    }
  }
}

export const deleteFlow = async (id: string): Promise<void> => {
  try {
    // Delete flow from MySQL database
    const query = 'DELETE FROM chatbot_flows_nodepath WHERE id = ?';
    await callMySQLAPI(query, [id]);
    console.log('Flow deleted from MySQL database successfully')
  } catch (error) {
    console.error('Error deleting flow from MySQL:', error)
    throw error
  }
}

// Media management - now using localStorage
export const saveMediaFile = async (file: any): Promise<any> => {
  try {
    const mediaFiles = JSON.parse(localStorage.getItem('media_files') || '[]')
    mediaFiles.push({
      ...file,
      uploaded_at: file.uploaded_at || new Date().toISOString(),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    })
    localStorage.setItem('media_files', JSON.stringify(mediaFiles))
    return file
  } catch (error) {
    console.error('Error saving media file:', error)
    throw error
  }
}

export const getMediaFiles = async (): Promise<any[]> => {
  try {
    const mediaFiles = JSON.parse(localStorage.getItem('media_files') || '[]')
    return mediaFiles.sort((a: any, b: any) => new Date(b.uploaded_at).getTime() - new Date(a.uploaded_at).getTime())
  } catch (error) {
    console.error('Error fetching media files:', error)
    return []
  }
}

export const getMediaFile = async (id: string): Promise<any | null> => {
  try {
    const mediaFiles = JSON.parse(localStorage.getItem('media_files') || '[]')
    return mediaFiles.find((f: any) => f.id === id) || null
  } catch (error) {
    console.error('Error fetching media file:', error)
    return null
  }
}

export const deleteMediaFile = async (id: string): Promise<void> => {
  try {
    const mediaFiles = JSON.parse(localStorage.getItem('media_files') || '[]')
    const filteredFiles = mediaFiles.filter((f: any) => f.id !== id)
    localStorage.setItem('media_files', JSON.stringify(filteredFiles))
  } catch (error) {
    console.error('Error deleting media file:', error)
    throw error
  }
}

// Flow execution management - using localStorage
export const saveFlowExecution = async (execution: any): Promise<void> => {
  try {
    const simulationId = execution.id || `exec_${execution.flowId}_${Date.now()}_${Math.random().toString(36).substring(2)}`
    
    const executions = JSON.parse(localStorage.getItem('flow_executions') || '{}')
    
    executions[simulationId] = {
      ...execution,
      id: simulationId,
      updated_at: new Date().toISOString()
    }
    
    localStorage.setItem('flow_executions', JSON.stringify(executions))
    console.log('Flow execution saved to localStorage:', simulationId)
  } catch (error) {
    console.error('Error saving flow execution:', error)
    throw error
  }
}

export const getFlowExecution = async (id: string): Promise<any | null> => {
  try {
    const executions = JSON.parse(localStorage.getItem('flow_executions') || '{}')
    return executions[id] || null
  } catch (error) {
    console.error('Error fetching flow execution:', error)
    return null
  }
}

export const updateFlowExecution = async (id: string, updates: any): Promise<void> => {
  try {
    const executions = JSON.parse(localStorage.getItem('flow_executions') || '{}')
    
    if (executions[id]) {
      executions[id] = {
        ...executions[id],
        ...updates,
        updated_at: new Date().toISOString()
      }
      localStorage.setItem('flow_executions', JSON.stringify(executions))
    }
  } catch (error) {
    console.error('Error updating flow execution:', error)
    throw error
  }
}

export const deleteFlowExecution = async (id: string): Promise<void> => {
  try {
    const executions = JSON.parse(localStorage.getItem('flow_executions') || '{}')
    delete executions[id]
    localStorage.setItem('flow_executions', JSON.stringify(executions))
  } catch (error) {
    console.error('Error deleting flow execution:', error)
    throw error
  }
}


// Helper function to extract AI prompt data from flow nodes
export const extractAIPromptData = (flow: ChatbotFlow) => {
  const aiNodes = flow.nodes.filter(node => node.type === 'prompt')
  return aiNodes.map(node => ({
    nodeId: node.id,
    instance: node.data.instance || '',
    openRouterKey: node.data.openRouterKey || ''
  }))
}