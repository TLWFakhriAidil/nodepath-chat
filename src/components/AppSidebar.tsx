import React from 'react';
import { NavLink } from 'react-router-dom';
import { 
  Bot, 
  Workflow, 
  MessageCircle, 
  FolderOpen, 
  Image,
  BarChart3,
  Settings,
  ChevronDown,
  ChevronRight,
  Home,
  Smartphone,
  Brain,
  User,
  HelpCircle
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
  SidebarMenuSub,
  SidebarMenuSubItem,
  SidebarMenuSubButton,
  useSidebar,
} from '@/components/ui/sidebar';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
import { useSidebarNavigation, createNavigationItem } from '@/hooks/useSidebarNavigation';

// Main navigation items
const mainItems = [
    createNavigationItem('Dashboard', Home, 'Overview and quick access', '/'),
    createNavigationItem('Flow Builder', GitBranch, 'Create and edit chatbot flows', '/'),
    createNavigationItem('Flow Manager', FolderOpen, 'Manage and preview saved flows', '/flows'),
    createNavigationItem('Test Chat', MessageSquare, 'Test your chatbot flows', '/test-chat'),
    createNavigationItem('Media Library', Image, 'Manage images, audio, and video files', '/media'),
    createNavigationItem('Lead Analytics', BarChart3, 'Track and analyze leads from chatbot conversations', '/analytics')
  ];

// Settings navigation items
const settingsItems = [
    createNavigationItem('WhatsApp', Smartphone, 'Configure WhatsApp integration', 'whatsapp-modal'),
    createNavigationItem('AI Settings', Brain, 'Configure AI model settings', 'ai-settings-modal'),
    createNavigationItem('Device Setting', Settings, 'Configure device and hardware settings', 'device-settings-modal'),
    createNavigationItem('Profile', User, 'Manage your profile settings', 'profile-modal'),
    createNavigationItem('Help & Support', HelpCircle, 'Get help and support', 'help-modal')
  ];

export function AppSidebar() {
  const { state } = useSidebar();
  const isCollapsed = state === 'collapsed';
  
  // Use the custom navigation hook
  const {
    expandedSections,
    isActive,
    getNavClasses,
    toggleSection,
    handleNavigationAction
  } = useSidebarNavigation([...mainItems, ...settingsItems]);

  const isSettingsOpen = expandedSections['settings'] || false;

  return (
    <Sidebar
      className={`fixed left-0 top-0 h-screen border-r bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 z-50 transition-all duration-300 ${isCollapsed ? "w-14" : "w-64"}`}
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

        {/* Main Navigation */}
        <SidebarGroup>
          <SidebarGroupLabel>Navigation</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {mainItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild className="h-auto p-3">
                    <NavLink to={item.url!} className={getNavClasses(item.url!)}>
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

        {/* Settings Section */}
        <SidebarGroup>
          <SidebarGroupLabel>Settings</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <Collapsible 
                open={isSettingsOpen} 
                onOpenChange={() => toggleSection('settings')}
                className="group/collapsible"
              >
                <SidebarMenuItem>
                  <CollapsibleTrigger asChild>
                    <SidebarMenuButton className="h-auto p-3">
                      <Settings className="mr-3 h-5 w-5 flex-shrink-0" />
                      {!isCollapsed && (
                        <>
                          <div className="flex-1">
                            <div className="font-medium">Settings</div>
                            <div className="text-xs text-muted-foreground">
                              Configure application settings
                            </div>
                          </div>
                          {isSettingsOpen ? (
                            <ChevronDown className="h-4 w-4" />
                          ) : (
                            <ChevronRight className="h-4 w-4" />
                          )}
                        </>
                      )}
                    </SidebarMenuButton>
                  </CollapsibleTrigger>
                  <CollapsibleContent>
                    <SidebarMenuSub>
                      {settingsItems.map((item) => (
                        <SidebarMenuSubItem key={item.title}>
                          <SidebarMenuSubButton
                            onClick={() => handleNavigationAction(item.action!)}
                            className="cursor-pointer"
                          >
                            <item.icon className="mr-2 h-4 w-4" />
                            {!isCollapsed && (
                              <div className="flex-1">
                                <div className="font-medium text-sm">{item.title}</div>
                                <div className="text-xs text-muted-foreground">
                                  {item.description}
                                </div>
                              </div>
                            )}
                          </SidebarMenuSubButton>
                        </SidebarMenuSubItem>
                      ))}
                    </SidebarMenuSub>
                  </CollapsibleContent>
                </SidebarMenuItem>
              </Collapsible>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  );
}