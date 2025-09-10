// NodePath Chat - Flow Builder JavaScript

class FlowBuilder {
    constructor(canvasId, toolboxId, propertiesId) {
        this.canvas = document.getElementById(canvasId);
        this.toolbox = document.getElementById(toolboxId);
        this.properties = document.getElementById(propertiesId);
        
        this.nodes = new Map();
        this.edges = new Map();
        this.selectedNode = null;
        this.selectedEdge = null;
        this.draggedNode = null;
        this.isConnecting = false;
        this.connectionStart = null;
        
        this.scale = 1;
        this.panX = 0;
        this.panY = 0;
        this.isPanning = false;
        this.lastPanPoint = { x: 0, y: 0 };
        
        this.currentFlow = null;
        this.isDirty = false;
        
        this.init();
    }
    
    init() {
        this.setupCanvas();
        this.setupToolbox();
        this.setupEventListeners();
        this.render();
    }
    
    setupCanvas() {
        this.canvas.style.position = 'relative';
        this.canvas.style.overflow = 'hidden';
        this.canvas.style.cursor = 'grab';
        this.canvas.style.background = `
            radial-gradient(circle, #e2e8f0 1px, transparent 1px)
        `;
        this.canvas.style.backgroundSize = '20px 20px';
    }
    
    setupToolbox() {
        const nodeTypes = [
            {
                type: 'ai_prompt',
                label: 'AI Prompt',
                icon: '🤖',
                color: '#4f46e5',
                description: 'AI-powered response node'
            },

            {
                type: 'condition',
                label: 'Condition',
                icon: '🔀',
                color: '#dc2626',
                description: 'Conditional branching'
            },
            {
                type: 'delay',
                label: 'Delay',
                icon: '⏱️',
                color: '#7c2d12',
                description: 'Add delay between messages'
            }
        ];
        
        this.toolbox.innerHTML = '';
        nodeTypes.forEach(nodeType => {
            const nodeElement = document.createElement('div');
            nodeElement.className = 'toolbox-node';
            nodeElement.draggable = true;
            nodeElement.dataset.nodeType = nodeType.type;
            nodeElement.style.cssText = `
                padding: 12px;
                margin: 8px 0;
                background: white;
                border: 2px solid ${nodeType.color};
                border-radius: 8px;
                cursor: grab;
                transition: all 0.2s ease;
                display: flex;
                align-items: center;
                gap: 8px;
            `;
            
            nodeElement.innerHTML = `
                <span style="font-size: 20px;">${nodeType.icon}</span>
                <div>
                    <div style="font-weight: 600; color: ${nodeType.color};">${nodeType.label}</div>
                    <div style="font-size: 12px; color: #6b7280;">${nodeType.description}</div>
                </div>
            `;
            
            nodeElement.addEventListener('dragstart', (e) => {
                e.dataTransfer.setData('text/plain', nodeType.type);
                nodeElement.style.opacity = '0.5';
            });
            
            nodeElement.addEventListener('dragend', () => {
                nodeElement.style.opacity = '1';
            });
            
            nodeElement.addEventListener('mouseenter', () => {
                nodeElement.style.transform = 'translateY(-2px)';
                nodeElement.style.boxShadow = '0 4px 12px rgba(0,0,0,0.15)';
            });
            
            nodeElement.addEventListener('mouseleave', () => {
                nodeElement.style.transform = 'translateY(0)';
                nodeElement.style.boxShadow = 'none';
            });
            
            this.toolbox.appendChild(nodeElement);
        });
    }
    
    setupEventListeners() {
        // Canvas events
        this.canvas.addEventListener('dragover', (e) => {
            e.preventDefault();
        });
        
        this.canvas.addEventListener('drop', (e) => {
            e.preventDefault();
            const nodeType = e.dataTransfer.getData('text/plain');
            const rect = this.canvas.getBoundingClientRect();
            const x = (e.clientX - rect.left - this.panX) / this.scale;
            const y = (e.clientY - rect.top - this.panY) / this.scale;
            
            this.addNode(nodeType, x, y);
        });
        
        this.canvas.addEventListener('click', (e) => {
            if (e.target === this.canvas) {
                this.selectNode(null);
            }
        });
        
        this.canvas.addEventListener('mousedown', (e) => {
            if (e.target === this.canvas) {
                this.isPanning = true;
                this.lastPanPoint = { x: e.clientX, y: e.clientY };
                this.canvas.style.cursor = 'grabbing';
            }
        });
        
        this.canvas.addEventListener('mousemove', (e) => {
            if (this.isPanning) {
                const deltaX = e.clientX - this.lastPanPoint.x;
                const deltaY = e.clientY - this.lastPanPoint.y;
                
                this.panX += deltaX;
                this.panY += deltaY;
                
                this.lastPanPoint = { x: e.clientX, y: e.clientY };
                this.render();
            }
        });
        
        this.canvas.addEventListener('mouseup', () => {
            this.isPanning = false;
            this.canvas.style.cursor = 'grab';
        });
        
        this.canvas.addEventListener('wheel', (e) => {
            e.preventDefault();
            const delta = e.deltaY > 0 ? 0.9 : 1.1;
            this.scale = Math.max(0.1, Math.min(3, this.scale * delta));
            this.render();
        });
        
        // Keyboard events
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Delete' && this.selectedNode) {
                this.deleteNode(this.selectedNode.id);
            }
            if (e.key === 'Escape') {
                this.selectNode(null);
                this.isConnecting = false;
                this.connectionStart = null;
            }
        });
    }
    
    addNode(type, x, y) {
        const nodeId = 'node_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
        
        const node = {
            id: nodeId,
            type: type,
            x: x,
            y: y,
            width: 200,
            height: 80,
            data: this.getDefaultNodeData(type)
        };
        
        this.nodes.set(nodeId, node);
        this.selectNode(node);
        this.markDirty();
        this.render();
        
        return node;
    }
    
    getDefaultNodeData(type) {
        switch (type) {
            case 'ai_prompt':
                return {
                    system_prompt: '',
                    instance: '',
                    apiprovider: '',
                    model: 'openai/gpt-4.1',
                    temperature: 0.7,
                    max_tokens: 1000
                };

            case 'condition':
                return {
                    variable: '',
                    operator: 'equals',
                    value: '',
                    true_path: '',
                    false_path: ''
                };
            case 'delay':
                return {
                    duration: 1000,
                    unit: 'milliseconds'
                };
            default:
                return {};
        }
    }
    
    deleteNode(nodeId) {
        if (this.nodes.has(nodeId)) {
            // Remove all edges connected to this node
            const edgesToRemove = [];
            this.edges.forEach((edge, edgeId) => {
                if (edge.source === nodeId || edge.target === nodeId) {
                    edgesToRemove.push(edgeId);
                }
            });
            
            edgesToRemove.forEach(edgeId => {
                this.edges.delete(edgeId);
            });
            
            this.nodes.delete(nodeId);
            
            if (this.selectedNode && this.selectedNode.id === nodeId) {
                this.selectNode(null);
            }
            
            this.markDirty();
            this.render();
        }
    }
    
    selectNode(node) {
        this.selectedNode = node;
        this.updatePropertiesPanel();
        this.render();
    }
    
    updatePropertiesPanel() {
        if (!this.selectedNode) {
            this.properties.innerHTML = `
                <div class="text-center text-muted py-4">
                    <i class="bi bi-cursor" style="font-size: 2rem;"></i>
                    <p class="mt-2">Select a node to edit its properties</p>
                </div>
            `;
            return;
        }
        
        const node = this.selectedNode;
        let propertiesHtml = `
            <div class="properties-header">
                <h6 class="mb-3">
                    <i class="bi bi-gear"></i> 
                    ${this.getNodeTypeLabel(node.type)} Properties
                </h6>
            </div>
            
            <div class="mb-3">
                <label class="form-label">Node ID</label>
                <input type="text" class="form-control" value="${node.id}" readonly>
            </div>
        `;
        
        switch (node.type) {
            case 'ai_prompt':
                propertiesHtml += this.getAIPromptProperties(node);
                break;
            case 'manual':
                propertiesHtml += this.getManualProperties(node);
                break;
            case 'condition':
                propertiesHtml += this.getConditionProperties(node);
                break;
            case 'delay':
                propertiesHtml += this.getDelayProperties(node);
                break;
        }
        
        propertiesHtml += `
            <div class="mt-4 pt-3 border-top">
                <button class="btn btn-danger btn-sm w-100" onclick="flowBuilder.deleteNode('${node.id}')">
                    <i class="bi bi-trash"></i> Delete Node
                </button>
            </div>
        `;
        
        this.properties.innerHTML = propertiesHtml;
        this.bindPropertyEvents();
    }
    
    getAIPromptProperties(node) {
        return `
            <div class="mb-3">
                <label class="form-label">System Prompt</label>
                <textarea class="form-control" rows="4" data-property="system_prompt" placeholder="Enter the system prompt for AI...">${node.data.system_prompt || ''}</textarea>
                <small class="form-text text-muted">Use {{variable}} for dynamic content</small>
            </div>
            
            <div class="mb-3">
                <label class="form-label">Instance</label>
                <input type="text" class="form-control" data-property="instance" value="${node.data.instance || ''}" placeholder="Instance identifier">
            </div>
            
            <div class="mb-3">
                <label class="form-label">OpenRouter API Key</label>
                <input type="password" class="form-control" data-property="apiprovider" value="${node.data.apiprovider || ''}" placeholder="sk-or-...">
            </div>
            
            <div class="mb-3">
                <label class="form-label">Model</label>
                <select class="form-select" data-property="model">
                    <option value="openai/gpt-4.1" ${node.data.model === 'openai/gpt-4.1' ? 'selected' : ''}>GPT-4.1</option>
                    <option value="openai/gpt-4" ${node.data.model === 'openai/gpt-4' ? 'selected' : ''}>GPT-4</option>
                    <option value="openai/gpt-3.5-turbo" ${node.data.model === 'openai/gpt-3.5-turbo' ? 'selected' : ''}>GPT-3.5 Turbo</option>
                </select>
            </div>
            
            <div class="row">
                <div class="col-6">
                    <label class="form-label">Temperature</label>
                    <input type="number" class="form-control" data-property="temperature" value="${node.data.temperature || 0.7}" min="0" max="2" step="0.1">
                </div>
                <div class="col-6">
                    <label class="form-label">Max Tokens</label>
                    <input type="number" class="form-control" data-property="max_tokens" value="${node.data.max_tokens || 1000}" min="1" max="4000">
                </div>
            </div>
        `;
    }
    
    getManualProperties(node) {
        return `
            <div class="mb-3">
                <label class="form-label">Content</label>
                <textarea class="form-control" rows="4" data-property="content" placeholder="Enter the manual response...">${node.data.content || ''}</textarea>
            </div>
            
            <div class="mb-3">
                <label class="form-label">Media Type</label>
                <select class="form-select" data-property="media_type">
                    <option value="text" ${node.data.media_type === 'text' ? 'selected' : ''}>Text Only</option>
                    <option value="image" ${node.data.media_type === 'image' ? 'selected' : ''}>Image</option>
                    <option value="audio" ${node.data.media_type === 'audio' ? 'selected' : ''}>Audio</option>
                    <option value="video" ${node.data.media_type === 'video' ? 'selected' : ''}>Video</option>
                    <option value="document" ${node.data.media_type === 'document' ? 'selected' : ''}>Document</option>
                </select>
            </div>
            
            <div class="mb-3" id="media-url-group" style="display: ${node.data.media_type !== 'text' ? 'block' : 'none'}">
                <label class="form-label">Media URL</label>
                <input type="url" class="form-control" data-property="media_url" value="${node.data.media_url || ''}" placeholder="https://...">
            </div>
        `;
    }
    
    getConditionProperties(node) {
        return `
            <div class="mb-3">
                <label class="form-label">Variable</label>
                <input type="text" class="form-control" data-property="variable" value="${node.data.variable || ''}" placeholder="Variable name">
            </div>
            
            <div class="mb-3">
                <label class="form-label">Operator</label>
                <select class="form-select" data-property="operator">
                    <option value="equals" ${node.data.operator === 'equals' ? 'selected' : ''}>Equals</option>
                    <option value="not_equals" ${node.data.operator === 'not_equals' ? 'selected' : ''}>Not Equals</option>
                    <option value="contains" ${node.data.operator === 'contains' ? 'selected' : ''}>Contains</option>
                    <option value="not_contains" ${node.data.operator === 'not_contains' ? 'selected' : ''}>Not Contains</option>
                    <option value="greater_than" ${node.data.operator === 'greater_than' ? 'selected' : ''}>Greater Than</option>
                    <option value="less_than" ${node.data.operator === 'less_than' ? 'selected' : ''}>Less Than</option>
                </select>
            </div>
            
            <div class="mb-3">
                <label class="form-label">Value</label>
                <input type="text" class="form-control" data-property="value" value="${node.data.value || ''}" placeholder="Comparison value">
            </div>
        `;
    }
    
    getDelayProperties(node) {
        return `
            <div class="row">
                <div class="col-8">
                    <label class="form-label">Duration</label>
                    <input type="number" class="form-control" data-property="duration" value="${node.data.duration || 1000}" min="100">
                </div>
                <div class="col-4">
                    <label class="form-label">Unit</label>
                    <select class="form-select" data-property="unit">
                        <option value="milliseconds" ${node.data.unit === 'milliseconds' ? 'selected' : ''}>ms</option>
                        <option value="seconds" ${node.data.unit === 'seconds' ? 'selected' : ''}>sec</option>
                        <option value="minutes" ${node.data.unit === 'minutes' ? 'selected' : ''}>min</option>
                    </select>
                </div>
            </div>
        `;
    }
    
    bindPropertyEvents() {
        const propertyInputs = this.properties.querySelectorAll('[data-property]');
        propertyInputs.forEach(input => {
            input.addEventListener('input', (e) => {
                const property = e.target.dataset.property;
                const value = e.target.type === 'number' ? parseFloat(e.target.value) : e.target.value;
                
                if (this.selectedNode) {
                    this.selectedNode.data[property] = value;
                    this.markDirty();
                    
                    // Special handling for media type
                    if (property === 'media_type') {
                        const mediaUrlGroup = document.getElementById('media-url-group');
                        if (mediaUrlGroup) {
                            mediaUrlGroup.style.display = value !== 'text' ? 'block' : 'none';
                        }
                    }
                }
            });
        });
    }
    
    getNodeTypeLabel(type) {
        const labels = {
            'ai_prompt': 'AI Prompt',
            'manual': 'Manual',
            'condition': 'Condition',
            'delay': 'Delay'
        };
        return labels[type] || type;
    }
    
    render() {
        // Clear canvas
        this.canvas.innerHTML = '';
        
        // Apply transform
        const transform = `translate(${this.panX}px, ${this.panY}px) scale(${this.scale})`;
        this.canvas.style.transform = transform;
        
        // Render edges first (so they appear behind nodes)
        this.edges.forEach(edge => {
            this.renderEdge(edge);
        });
        
        // Render nodes
        this.nodes.forEach(node => {
            this.renderNode(node);
        });
    }
    
    renderNode(node) {
        const nodeElement = document.createElement('div');
        nodeElement.className = 'flow-node';
        nodeElement.style.cssText = `
            position: absolute;
            left: ${node.x}px;
            top: ${node.y}px;
            width: ${node.width}px;
            height: ${node.height}px;
            background: white;
            border: 2px solid ${this.getNodeColor(node.type)};
            border-radius: 8px;
            padding: 12px;
            cursor: pointer;
            user-select: none;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            transition: all 0.2s ease;
            ${this.selectedNode && this.selectedNode.id === node.id ? 'border-color: #4f46e5; box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.2);' : ''}
        `;
        
        const icon = this.getNodeIcon(node.type);
        const label = this.getNodeTypeLabel(node.type);
        const content = this.getNodeContent(node);
        
        nodeElement.innerHTML = `
            <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">
                <span style="font-size: 16px;">${icon}</span>
                <span style="font-weight: 600; color: ${this.getNodeColor(node.type)};">${label}</span>
            </div>
            <div style="font-size: 12px; color: #6b7280; overflow: hidden; text-overflow: ellipsis;">
                ${content}
            </div>
        `;
        
        // Add event listeners
        nodeElement.addEventListener('click', (e) => {
            e.stopPropagation();
            this.selectNode(node);
        });
        
        nodeElement.addEventListener('mousedown', (e) => {
            e.stopPropagation();
            this.draggedNode = node;
            this.dragOffset = {
                x: e.clientX - node.x * this.scale,
                y: e.clientY - node.y * this.scale
            };
        });
        
        this.canvas.appendChild(nodeElement);
    }
    
    renderEdge(edge) {
        const sourceNode = this.nodes.get(edge.source);
        const targetNode = this.nodes.get(edge.target);
        
        if (!sourceNode || !targetNode) return;
        
        const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        svg.style.cssText = `
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            pointer-events: none;
            z-index: -1;
        `;
        
        const line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        line.setAttribute('x1', sourceNode.x + sourceNode.width / 2);
        line.setAttribute('y1', sourceNode.y + sourceNode.height);
        line.setAttribute('x2', targetNode.x + targetNode.width / 2);
        line.setAttribute('y2', targetNode.y);
        line.setAttribute('stroke', '#6b7280');
        line.setAttribute('stroke-width', '2');
        line.setAttribute('marker-end', 'url(#arrowhead)');
        
        // Add arrowhead marker
        const defs = document.createElementNS('http://www.w3.org/2000/svg', 'defs');
        const marker = document.createElementNS('http://www.w3.org/2000/svg', 'marker');
        marker.setAttribute('id', 'arrowhead');
        marker.setAttribute('markerWidth', '10');
        marker.setAttribute('markerHeight', '7');
        marker.setAttribute('refX', '9');
        marker.setAttribute('refY', '3.5');
        marker.setAttribute('orient', 'auto');
        
        const polygon = document.createElementNS('http://www.w3.org/2000/svg', 'polygon');
        polygon.setAttribute('points', '0 0, 10 3.5, 0 7');
        polygon.setAttribute('fill', '#6b7280');
        
        marker.appendChild(polygon);
        defs.appendChild(marker);
        svg.appendChild(defs);
        svg.appendChild(line);
        
        this.canvas.appendChild(svg);
    }
    
    getNodeColor(type) {
        const colors = {
            'ai_prompt': '#4f46e5',
            'manual': '#059669',
            'condition': '#dc2626',
            'delay': '#7c2d12'
        };
        return colors[type] || '#6b7280';
    }
    
    getNodeIcon(type) {
        const icons = {
            'ai_prompt': '🤖',
            'manual': '💬',
            'condition': '🔀',
            'delay': '⏱️'
        };
        return icons[type] || '📄';
    }
    
    getNodeContent(node) {
        switch (node.type) {
            case 'ai_prompt':
                return node.data.system_prompt ? 
                    node.data.system_prompt.substring(0, 50) + '...' : 
                    'No prompt configured';
            case 'manual':
                return node.data.content ? 
                    node.data.content.substring(0, 50) + '...' : 
                    'No content configured';
            case 'condition':
                return node.data.variable ? 
                    `${node.data.variable} ${node.data.operator} ${node.data.value}` : 
                    'No condition configured';
            case 'delay':
                return `${node.data.duration} ${node.data.unit}`;
            default:
                return 'Node';
        }
    }
    
    markDirty() {
        this.isDirty = true;
        // Update save button state
        const saveBtn = document.getElementById('save-flow');
        if (saveBtn) {
            saveBtn.textContent = 'Save Flow *';
            saveBtn.classList.add('btn-warning');
            saveBtn.classList.remove('btn-primary');
        }
    }
    
    markClean() {
        this.isDirty = false;
        const saveBtn = document.getElementById('save-flow');
        if (saveBtn) {
            saveBtn.textContent = 'Save Flow';
            saveBtn.classList.add('btn-primary');
            saveBtn.classList.remove('btn-warning');
        }
    }
    
    exportFlow() {
        const nodes = Array.from(this.nodes.values());
        const edges = Array.from(this.edges.values());
        
        return {
            nodes: nodes,
            edges: edges,
            metadata: {
                version: '1.0.0',
                created: new Date().toISOString(),
                scale: this.scale,
                pan: { x: this.panX, y: this.panY }
            }
        };
    }
    
    importFlow(flowData) {
        this.nodes.clear();
        this.edges.clear();
        
        if (flowData.nodes) {
            flowData.nodes.forEach(node => {
                this.nodes.set(node.id, node);
            });
        }
        
        if (flowData.edges) {
            flowData.edges.forEach(edge => {
                this.edges.set(edge.id, edge);
            });
        }
        
        if (flowData.metadata) {
            this.scale = flowData.metadata.scale || 1;
            this.panX = flowData.metadata.pan?.x || 0;
            this.panY = flowData.metadata.pan?.y || 0;
        }
        
        this.selectNode(null);
        this.markClean();
        this.render();
    }
    
    clearCanvas() {
        this.nodes.clear();
        this.edges.clear();
        this.selectNode(null);
        this.scale = 1;
        this.panX = 0;
        this.panY = 0;
        this.markClean();
        this.render();
    }
}

// Export for global use
if (typeof window !== 'undefined') {
    window.FlowBuilder = FlowBuilder;
}