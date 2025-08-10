import { useState, useCallback, useEffect } from 'react';
import { useLocation } from 'react-router-dom';

export interface NavigationItem {
  title: string;
  url?: string;
  icon: React.ComponentType<any>;
  description: string;
  action?: string;
  children?: NavigationItem[];
}

export interface SidebarNavigationState {
  expandedSections: Record<string, boolean>;
  activeItem: string | null;
}

export function useSidebarNavigation(items: NavigationItem[]) {
  const location = useLocation();
  const [expandedSections, setExpandedSections] = useState<Record<string, boolean>>({});
  const [activeItem, setActiveItem] = useState<string | null>(null);

  // Determine active item based on current route
  const isActive = useCallback((path: string) => {
    if (path === '/' && location.pathname === '/') return true;
    if (path !== '/' && location.pathname.startsWith(path)) return true;
    return false;
  }, [location.pathname]);

  // Get navigation classes for styling
  const getNavClasses = useCallback((path: string) => {
    const active = isActive(path);
    return active 
      ? "bg-primary/10 text-primary font-medium border-l-2 border-primary" 
      : "hover:bg-muted/50 text-foreground";
  }, [isActive]);

  // Toggle section expansion
  const toggleSection = useCallback((sectionKey: string) => {
    setExpandedSections(prev => ({
      ...prev,
      [sectionKey]: !prev[sectionKey]
    }));
  }, []);

  // Handle navigation actions
  const handleNavigationAction = useCallback((action: string) => {
    switch (action) {
      case 'whatsapp-modal':
        // Trigger WhatsApp modal
        const whatsappModal = document.getElementById('whatsappModal');
        if (whatsappModal) {
          console.log('Opening WhatsApp settings modal');
          // You can dispatch custom events or use a modal library here
          window.dispatchEvent(new CustomEvent('open-whatsapp-modal'));
        }
        break;
      case 'ai-settings-modal':
        // Trigger AI Settings modal
        const aiModal = document.getElementById('aiSettingsModal');
        if (aiModal) {
          console.log('Opening AI settings modal');
          window.dispatchEvent(new CustomEvent('open-ai-settings-modal'));
        }
        break;
      case 'device-settings-modal':
        // Dispatch custom event for device settings modal
        window.dispatchEvent(new CustomEvent('open-device-settings-modal'));
        console.log('Opening device settings...');
        break;
      case 'profile-settings':
        // Navigate to profile settings or open modal
        console.log('Opening profile settings');
        window.dispatchEvent(new CustomEvent('open-profile-settings'));
        break;
      case 'help-support':
        // Navigate to help page or open support modal
        console.log('Opening help and support');
        window.dispatchEvent(new CustomEvent('open-help-support'));
        break;
      default:
        console.warn(`Unknown navigation action: ${action}`);
        break;
    }
  }, []);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      // Ctrl/Cmd + Shift + S to toggle settings
      if ((event.ctrlKey || event.metaKey) && event.shiftKey && event.key === 'S') {
        event.preventDefault();
        toggleSection('settings');
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [toggleSection]);

  // Update active item when location changes
  useEffect(() => {
    const currentItem = items.find(item => item.url && isActive(item.url));
    setActiveItem(currentItem?.title || null);
  }, [location.pathname, items, isActive]);

  return {
    expandedSections,
    activeItem,
    isActive,
    getNavClasses,
    toggleSection,
    handleNavigationAction,
    setExpandedSections
  };
}

// Utility function to create navigation items
export function createNavigationItem(
  title: string,
  icon: React.ComponentType<any>,
  description: string,
  urlOrAction?: string,
  children?: NavigationItem[]
): NavigationItem {
  const isUrl = urlOrAction?.startsWith('/');
  return {
    title,
    icon,
    description,
    ...(isUrl ? { url: urlOrAction } : { action: urlOrAction }),
    children
  };
}

// Export types for external use
export type { NavigationItem, SidebarNavigationState };