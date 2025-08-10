# Enhanced Sidebar Navigation System

This document explains the enhanced sidebar navigation system that has been implemented in the application.

## Features

### 🎯 Core Features
- **Fixed Left Navigation**: Sidebar is positioned as a fixed left panel
- **Collapsible Navigation**: Sidebar can be collapsed to icon-only mode (64px to 56px width)
- **Active State Management**: Automatically highlights the current page
- **Expandable Settings Section**: Settings section with collapsible sub-items
- **Keyboard Shortcuts**: `Ctrl/Cmd + Shift + S` to toggle settings section
- **Custom Actions**: Support for modal triggers and custom actions
- **Responsive Layout**: Main content area adjusts margin based on sidebar state
- **Smooth Transitions**: 300ms transition animations for all state changes

### 📱 Navigation Items

#### Main Navigation
- **Dashboard** - Overview and quick access
- **Flow Builder** - Create and edit chatbot flows
- **Flow Manager** - Manage and preview saved flows
- **Test Chat** - Test your chatbot flows
- **Media Library** - Manage images, audio, and video files
- **Lead Analytics** - Track and analyze leads from chatbot conversations

#### Settings Section
- **WhatsApp** - Configure WhatsApp integration
- **AI Settings** - Configure AI model settings
- **Device Setting** - Configure device and hardware settings
- **Profile** - Manage your profile settings
- **Help & Support** - Get help and support

## Usage

### Basic Implementation

```tsx
import { AppSidebar } from '@/components/AppSidebar';
import { SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar';

function App() {
  return (
    <SidebarProvider>
      <div className="min-h-screen flex w-full bg-background">
        <AppSidebar />
        <div className="flex-1 flex flex-col">
          <header className="h-12 flex items-center border-b">
            <SidebarTrigger className="ml-4" />
          </header>
          <main className="flex-1">
            {/* Your content here */}
          </main>
        </div>
      </div>
    </SidebarProvider>
  );
}
```

### Custom Navigation Hook

The `useSidebarNavigation` hook provides reusable navigation logic:

```tsx
import { useSidebarNavigation, createNavigationItem } from '@/hooks/useSidebarNavigation';
import { Home, Settings } from 'lucide-react';

const navigationItems = [
  createNavigationItem('Home', Home, 'Go to homepage', '/'),
  createNavigationItem('Settings', Settings, 'Open settings', 'settings-modal')
];

function MyComponent() {
  const {
    expandedSections,
    isActive,
    getNavClasses,
    toggleSection,
    handleNavigationAction
  } = useSidebarNavigation(navigationItems);

  // Use the navigation logic in your component
}
```

### Creating Navigation Items

```tsx
import { createNavigationItem } from '@/hooks/useSidebarNavigation';
import { Home, Settings } from 'lucide-react';

// For URL navigation
const homeItem = createNavigationItem(
  'Home',           // title
  Home,            // icon component
  'Go to homepage', // description
  '/'              // URL
);

// For custom actions
const settingsItem = createNavigationItem(
  'Settings',
  Settings,
  'Open settings modal',
  'settings-modal'  // action name
);
```

## Customization

### Adding New Navigation Items

1. **Main Navigation**: Add to `mainItems` array in `AppSidebar.tsx`
2. **Settings**: Add to `settingsItems` array in `AppSidebar.tsx`

```tsx
const mainItems = [
  // existing items...
  createNavigationItem('New Page', NewIcon, 'Description', '/new-page')
];
```

### Custom Actions

Add new actions to the `handleNavigationAction` function in `useSidebarNavigation.ts`:

```tsx
case 'my-custom-action':
  // Your custom logic here
  console.log('Custom action triggered');
  // Dispatch events, open modals, etc.
  break;
```

### Event Handling

The navigation system dispatches custom events for modal triggers:

```tsx
// Listen for navigation events
window.addEventListener('open-whatsapp-modal', () => {
  // Handle WhatsApp modal opening
});

window.addEventListener('open-ai-settings-modal', () => {
  // Handle AI settings modal opening
});
```

## Styling

### Active States
The navigation automatically applies active styles based on the current route:
- Active: `bg-primary/10 text-primary font-medium border-l-2 border-primary`
- Inactive: `hover:bg-muted/50 text-foreground`

### Collapsed State
When collapsed, the sidebar shows only icons and hides text content.

### Responsive Behavior
- **Desktop**: Fixed sidebar with collapse functionality
- **Mobile**: Overlay sidebar that can be toggled

## Keyboard Shortcuts

- `Ctrl/Cmd + Shift + S`: Toggle settings section
- `B`: Toggle sidebar (built-in shortcut)

## Technical Details

### Dependencies
- `@radix-ui/react-collapsible`: For collapsible sections
- `lucide-react`: For icons
- `react-router-dom`: For navigation
- Custom UI components from `@/components/ui/sidebar`

### File Structure
```
src/
├── components/
│   ├── AppSidebar.tsx          # Main sidebar component
│   └── ui/
│       ├── sidebar.tsx         # Base sidebar components
│       └── collapsible.tsx     # Collapsible functionality
└── hooks/
    └── useSidebarNavigation.ts # Navigation logic hook
```

### State Management
The navigation state is managed through:
- `useSidebar()`: Built-in sidebar state (collapsed/expanded)
- `useSidebarNavigation()`: Custom navigation state (active items, expanded sections)

## Best Practices

1. **Consistent Icons**: Use icons from the same icon library (lucide-react)
2. **Descriptive Text**: Provide clear descriptions for each navigation item
3. **Logical Grouping**: Group related items in the same section
4. **Accessibility**: Ensure keyboard navigation works properly
5. **Performance**: Use the custom hook to avoid prop drilling

## Troubleshooting

### Common Issues

1. **Navigation not highlighting**: Check if the URL matches the `isActive` logic
2. **Icons not showing**: Ensure icons are imported from lucide-react
3. **Collapsible not working**: Verify Collapsible components are imported correctly
4. **Actions not triggering**: Check the action name matches in `handleNavigationAction`

### Debug Tips

```tsx
// Enable debug logging in useSidebarNavigation.ts
console.log('Current path:', location.pathname);
console.log('Active item:', activeItem);
console.log('Expanded sections:', expandedSections);
```

This enhanced sidebar navigation system provides a robust, extensible foundation for application navigation with modern UX patterns and accessibility features.