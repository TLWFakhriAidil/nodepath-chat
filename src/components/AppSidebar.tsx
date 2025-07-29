import React from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { 
  Bot, 
  Workflow, 
  MessageCircle, 
  FolderOpen, 
  Image,
  Settings
} from 'lucide-react';
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarTrigger,
  useSidebar,
} from '@/components/ui/sidebar';

const mainItems = [
  { 
    title: 'Flow Builder', 
    url: '/', 
    icon: Workflow,
    description: 'Create and edit chatbot flows'
  },
  { 
    title: 'Flow Manager', 
    url: '/flows', 
    icon: FolderOpen,
    description: 'Manage and preview saved flows'
  },
  { 
    title: 'Test Chat', 
    url: '/test', 
    icon: MessageCircle,
    description: 'Test your chatbot flows'
  },
  { 
    title: 'Media Library', 
    url: '/media', 
    icon: Image,
    description: 'Manage images, audio, and video files'
  }
];

export function AppSidebar() {
  const { state } = useSidebar();
  const location = useLocation();
  const currentPath = location.pathname;
  const isCollapsed = state === 'collapsed';

  const isActive = (path: string) => {
    if (path === '/' && currentPath === '/') return true;
    if (path !== '/' && currentPath.startsWith(path)) return true;
    return false;
  };

  const getNavCls = (path: string) => {
    const active = isActive(path);
    return active 
      ? "bg-primary/10 text-primary font-medium border-l-2 border-primary" 
      : "hover:bg-muted/50 text-foreground";
  };

  return (
    <Sidebar
      className={isCollapsed ? "w-14" : "w-64"}
      collapsible="icon"
    >
      <SidebarContent>
        {/* Header */}
        <div className="p-4 border-b">
          <div className="flex items-center gap-2">
            <Bot className="w-6 h-6 text-primary" />
            {!isCollapsed && (
              <div>
                <h1 className="font-semibold text-lg">Chatbot Studio</h1>
                <p className="text-xs text-muted-foreground">Build conversational flows</p>
              </div>
            )}
          </div>
        </div>

        <SidebarGroup>
          <SidebarGroupLabel>Navigation</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {mainItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild className="h-auto p-3">
                    <NavLink to={item.url} className={getNavCls(item.url)}>
                      <item.icon className="mr-3 h-5 w-5 flex-shrink-0" />
                      {!isCollapsed && (
                        <div className="flex-1">
                          <div className="font-medium">{item.title}</div>
                          <div className="text-xs text-muted-foreground">
                            {item.description}
                          </div>
                        </div>
                      )}
                    </NavLink>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  );
}