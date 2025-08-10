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
  HelpCircle,
  GitBranch,
  MessageSquare
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
    createNavigationItem('Flow Builder', GitBranch, 'Create and edit chatbot flows', '/flow-builder'),
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
      className={`fixed left-0 top-0 h-screen border-r border-border/50 bg-card/95 backdrop-blur-xl supports-[backdrop-filter]:bg-card/90 z-50 transition-all duration-300 futuristic-border ${isCollapsed ? "w-14" : "w-56"}`}
      collapsible="icon"
    >
      <SidebarContent className="bg-transparent">
        {/* Header */}
        <div className="p-3 border-b border-border/50 bg-gradient-to-r from-primary/10 to-blue-600/10">
          <div className="flex items-center gap-2">
            <div className="relative">
              <Bot className="w-6 h-6 text-primary drop-shadow-lg" />
              <div className="absolute -top-1 -right-1 w-2 h-2 bg-primary rounded-full pulse-glow"></div>
            </div>
            {!isCollapsed && (
              <div>
                <h1 className="font-bold text-base text-foreground holographic-text">Chatbot Studio</h1>
                <p className="text-xs text-muted-foreground/80">Build conversational flows</p>
              </div>
            )}
          </div>
        </div>

        {/* Main Navigation */}
        <SidebarGroup>
          <SidebarGroupLabel className="text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-2 flex items-center gap-2">
            <div className="w-2 h-2 bg-primary rounded-full pulse-glow"></div>
            Navigation
          </SidebarGroupLabel>
          <SidebarGroupContent className="px-2">
            <SidebarMenu className="space-y-1">
              {mainItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild className="h-auto p-0">
                    <NavLink 
                      to={item.url!} 
                      className={`${getNavClasses(item.url!)} p-3 rounded-lg transition-all duration-300 hover:bg-primary/10 hover:border-primary/20 border border-transparent glow-on-hover group relative overflow-hidden`}
                    >
                      <item.icon className="mr-3 h-5 w-5 flex-shrink-0 text-primary group-hover:scale-110 transition-transform duration-300" />
                      {!isCollapsed && (
                        <div className="flex-1 relative z-10">
                          <div className="font-medium text-foreground group-hover:text-primary transition-colors">{item.title}</div>
                          <div className="text-xs text-muted-foreground/70 group-hover:text-muted-foreground transition-colors">
                            {item.description}
                          </div>
                        </div>
                      )}
                      <div className="absolute inset-0 bg-gradient-to-r from-transparent via-primary/5 to-transparent -translate-x-full group-hover:translate-x-full transition-transform duration-700"></div>
                    </NavLink>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        {/* Settings Section */}
        <SidebarGroup>
          <SidebarGroupLabel className="text-xs font-semibold text-muted-foreground uppercase tracking-wider px-4 py-2 flex items-center gap-2">
            <div className="w-2 h-2 bg-orange-500 rounded-full pulse-glow"></div>
            Settings
          </SidebarGroupLabel>
          <SidebarGroupContent className="px-2">
            <SidebarMenu>
              <Collapsible 
                open={isSettingsOpen} 
                onOpenChange={() => toggleSection('settings')}
                className="group/collapsible"
              >
                <SidebarMenuItem>
                  <CollapsibleTrigger asChild>
                    <SidebarMenuButton className="h-auto p-3 rounded-lg transition-all duration-300 hover:bg-orange-500/10 hover:border-orange-500/20 border border-transparent glow-on-hover group relative overflow-hidden">
                      <Settings className="mr-3 h-5 w-5 flex-shrink-0 text-orange-500 group-hover:scale-110 transition-transform duration-300" />
                      {!isCollapsed && (
                        <>
                          <div className="flex-1 relative z-10">
                            <div className="font-medium text-foreground group-hover:text-orange-500 transition-colors">Settings</div>
                            <div className="text-xs text-muted-foreground/70 group-hover:text-muted-foreground transition-colors">
                              Configure application settings
                            </div>
                          </div>
                          {isSettingsOpen ? (
                            <ChevronDown className="h-4 w-4 text-orange-500 transition-transform duration-300" />
                          ) : (
                            <ChevronRight className="h-4 w-4 text-orange-500 transition-transform duration-300" />
                          )}
                        </>
                      )}
                      <div className="absolute inset-0 bg-gradient-to-r from-transparent via-orange-500/5 to-transparent -translate-x-full group-hover:translate-x-full transition-transform duration-700"></div>
                    </SidebarMenuButton>
                  </CollapsibleTrigger>
                  <CollapsibleContent className="transition-all duration-300">
                    <SidebarMenuSub className="space-y-1 mt-2">
                      {settingsItems.map((item) => (
                        <SidebarMenuSubItem key={item.title}>
                          <SidebarMenuSubButton
                            onClick={() => handleNavigationAction(item.action!)}
                            className="cursor-pointer h-auto p-3 rounded-lg transition-all duration-300 hover:bg-orange-500/10 hover:border-orange-500/20 border border-transparent glow-on-hover group relative overflow-hidden ml-4"
                          >
                            <item.icon className="mr-3 h-4 w-4 text-orange-500 group-hover:scale-110 transition-transform duration-300" />
                            {!isCollapsed && (
                              <div className="flex-1 relative z-10">
                                <div className="font-medium text-sm text-foreground group-hover:text-orange-500 transition-colors">{item.title}</div>
                                <div className="text-xs text-muted-foreground/70 group-hover:text-muted-foreground transition-colors">
                                  {item.description}
                                </div>
                              </div>
                            )}
                            <div className="absolute inset-0 bg-gradient-to-r from-transparent via-orange-500/5 to-transparent -translate-x-full group-hover:translate-x-full transition-transform duration-700"></div>
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