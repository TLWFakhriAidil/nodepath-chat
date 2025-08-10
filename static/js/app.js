// NodePath Chat - Main JavaScript Application

// Global application state
window.NodePathChat = {
    config: {
        apiBaseUrl: '/api',
        wsBaseUrl: window.location.protocol === 'https:' ? 'wss://' : 'ws://' + window.location.host,
        version: '1.0.0'
    },
    state: {
        user: null,
        flows: [],
        currentExecution: null,
        wsConnection: null
    },
    utils: {},
    components: {}
};

// Utility Functions
NodePathChat.utils = {
    // API Helper Functions
    api: {
        get: function(endpoint, options = {}) {
            return fetch(NodePathChat.config.apiBaseUrl + endpoint, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                },
                ...options
            }).then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.json();
            });
        },
        
        post: function(endpoint, data = {}, options = {}) {
            return fetch(NodePathChat.config.apiBaseUrl + endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                },
                body: JSON.stringify(data),
                ...options
            }).then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.json();
            });
        },
        
        put: function(endpoint, data = {}, options = {}) {
            return fetch(NodePathChat.config.apiBaseUrl + endpoint, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                },
                body: JSON.stringify(data),
                ...options
            }).then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.json();
            });
        },
        
        delete: function(endpoint, options = {}) {
            return fetch(NodePathChat.config.apiBaseUrl + endpoint, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                },
                ...options
            }).then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.json();
            });
        }
    },
    
    // WebSocket Helper
    websocket: {
        connect: function(endpoint, onMessage, onError, onClose) {
            const wsUrl = NodePathChat.config.wsBaseUrl + endpoint;
            const ws = new WebSocket(wsUrl);
            
            ws.onopen = function() {
                console.log('WebSocket connected to:', wsUrl);
                NodePathChat.state.wsConnection = ws;
            };
            
            ws.onmessage = function(event) {
                try {
                    const data = JSON.parse(event.data);
                    if (onMessage) onMessage(data);
                } catch (e) {
                    console.error('Failed to parse WebSocket message:', e);
                }
            };
            
            ws.onerror = function(error) {
                console.error('WebSocket error:', error);
                if (onError) onError(error);
            };
            
            ws.onclose = function() {
                console.log('WebSocket connection closed');
                NodePathChat.state.wsConnection = null;
                if (onClose) onClose();
            };
            
            return ws;
        },
        
        send: function(data) {
            if (NodePathChat.state.wsConnection && NodePathChat.state.wsConnection.readyState === WebSocket.OPEN) {
                NodePathChat.state.wsConnection.send(JSON.stringify(data));
                return true;
            }
            return false;
        },
        
        disconnect: function() {
            if (NodePathChat.state.wsConnection) {
                NodePathChat.state.wsConnection.close();
                NodePathChat.state.wsConnection = null;
            }
        }
    },
    
    // UI Helper Functions
    ui: {
        showToast: function(title, message, type = 'info', duration = 5000) {
            // Create toast container if it doesn't exist
            let toastContainer = document.getElementById('toast-container');
            if (!toastContainer) {
                toastContainer = document.createElement('div');
                toastContainer.id = 'toast-container';
                toastContainer.className = 'toast-container position-fixed top-0 end-0 p-3';
                toastContainer.style.zIndex = '9999';
                document.body.appendChild(toastContainer);
            }
            
            // Create toast element
            const toastId = 'toast-' + Date.now();
            const toastHtml = `
                <div id="${toastId}" class="toast align-items-center text-white bg-${type} border-0" role="alert">
                    <div class="d-flex">
                        <div class="toast-body">
                            <strong>${title}</strong><br>
                            ${message}
                        </div>
                        <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
                    </div>
                </div>
            `;
            
            toastContainer.insertAdjacentHTML('beforeend', toastHtml);
            
            // Initialize and show toast
            const toastElement = document.getElementById(toastId);
            const toast = new bootstrap.Toast(toastElement, {
                autohide: true,
                delay: duration
            });
            
            toast.show();
            
            // Remove toast element after it's hidden
            toastElement.addEventListener('hidden.bs.toast', function() {
                toastElement.remove();
            });
        },
        
        showModal: function(title, content, buttons = []) {
            // Create modal if it doesn't exist
            let modal = document.getElementById('dynamic-modal');
            if (!modal) {
                const modalHtml = `
                    <div class="modal fade" id="dynamic-modal" tabindex="-1">
                        <div class="modal-dialog">
                            <div class="modal-content">
                                <div class="modal-header">
                                    <h5 class="modal-title"></h5>
                                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                                </div>
                                <div class="modal-body"></div>
                                <div class="modal-footer"></div>
                            </div>
                        </div>
                    </div>
                `;
                document.body.insertAdjacentHTML('beforeend', modalHtml);
                modal = document.getElementById('dynamic-modal');
            }
            
            // Set modal content
            modal.querySelector('.modal-title').textContent = title;
            modal.querySelector('.modal-body').innerHTML = content;
            
            // Set modal buttons
            const footer = modal.querySelector('.modal-footer');
            footer.innerHTML = '';
            buttons.forEach(button => {
                const btn = document.createElement('button');
                btn.type = 'button';
                btn.className = `btn btn-${button.type || 'secondary'}`;
                btn.textContent = button.text;
                if (button.dismiss) {
                    btn.setAttribute('data-bs-dismiss', 'modal');
                }
                if (button.onClick) {
                    btn.addEventListener('click', button.onClick);
                }
                footer.appendChild(btn);
            });
            
            // Show modal
            const bsModal = new bootstrap.Modal(modal);
            bsModal.show();
            
            return bsModal;
        },
        
        showLoading: function(element, show = true) {
            if (show) {
                element.innerHTML = `
                    <div class="d-flex justify-content-center align-items-center py-4">
                        <div class="spinner-border text-primary" role="status">
                            <span class="visually-hidden">Loading...</span>
                        </div>
                    </div>
                `;
            }
        },
        
        formatDate: function(dateString) {
            if (!dateString) return 'Unknown';
            const date = new Date(dateString);
            return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
        },
        
        formatFileSize: function(bytes) {
            if (bytes === 0) return '0 Bytes';
            const k = 1024;
            const sizes = ['Bytes', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        },
        
        escapeHtml: function(text) {
            if (typeof text !== 'string') return text;
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        },
        
        debounce: function(func, wait) {
            let timeout;
            return function executedFunction(...args) {
                const later = () => {
                    clearTimeout(timeout);
                    func(...args);
                };
                clearTimeout(timeout);
                timeout = setTimeout(later, wait);
            };
        }
    },
    
    // Validation Functions
    validation: {
        isEmail: function(email) {
            const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            return re.test(email);
        },
        
        isPhone: function(phone) {
            const re = /^[\+]?[1-9][\d]{0,15}$/;
            return re.test(phone.replace(/\s/g, ''));
        },
        
        isUrl: function(url) {
            try {
                new URL(url);
                return true;
            } catch {
                return false;
            }
        },
        
        isEmpty: function(value) {
            return value === null || value === undefined || value === '' || (Array.isArray(value) && value.length === 0);
        },
        
        isValidFlowName: function(name) {
            return name && name.trim().length >= 3 && name.trim().length <= 100;
        },
        
        isValidNodeId: function(nodeId) {
            const re = /^[a-zA-Z0-9_-]+$/;
            return nodeId && re.test(nodeId) && nodeId.length <= 50;
        }
    },
    
    // Storage Functions
    storage: {
        set: function(key, value) {
            try {
                localStorage.setItem(key, JSON.stringify(value));
                return true;
            } catch (e) {
                console.error('Failed to save to localStorage:', e);
                return false;
            }
        },
        
        get: function(key, defaultValue = null) {
            try {
                const item = localStorage.getItem(key);
                return item ? JSON.parse(item) : defaultValue;
            } catch (e) {
                console.error('Failed to read from localStorage:', e);
                return defaultValue;
            }
        },
        
        remove: function(key) {
            try {
                localStorage.removeItem(key);
                return true;
            } catch (e) {
                console.error('Failed to remove from localStorage:', e);
                return false;
            }
        },
        
        clear: function() {
            try {
                localStorage.clear();
                return true;
            } catch (e) {
                console.error('Failed to clear localStorage:', e);
                return false;
            }
        }
    }
};

// Global Error Handler
window.addEventListener('error', function(event) {
    console.error('Global error:', event.error);
    NodePathChat.utils.ui.showToast(
        'Error',
        'An unexpected error occurred. Please try again.',
        'danger'
    );
});

// Global Unhandled Promise Rejection Handler
window.addEventListener('unhandledrejection', function(event) {
    console.error('Unhandled promise rejection:', event.reason);
    NodePathChat.utils.ui.showToast(
        'Error',
        'A network or processing error occurred. Please try again.',
        'danger'
    );
});

// Initialize Application
document.addEventListener('DOMContentLoaded', function() {
    console.log('NodePath Chat initialized');
    
    // Initialize tooltips
    const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    tooltipTriggerList.map(function(tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl);
    });
    
    // Initialize popovers
    const popoverTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="popover"]'));
    popoverTriggerList.map(function(popoverTriggerEl) {
        return new bootstrap.Popover(popoverTriggerEl);
    });
    
    // Set active navigation item
    const currentPath = window.location.pathname;
    const navLinks = document.querySelectorAll('.nav-link');
    navLinks.forEach(link => {
        if (link.getAttribute('href') === currentPath) {
            link.classList.add('active');
        } else {
            link.classList.remove('active');
        }
    });
    
    // Auto-refresh functionality for dashboard
    if (currentPath === '/' || currentPath === '/analytics') {
        setInterval(function() {
            // Refresh data every 30 seconds
            const refreshEvent = new CustomEvent('autoRefresh');
            document.dispatchEvent(refreshEvent);
        }, 30000);
    }
});

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = NodePathChat;
}