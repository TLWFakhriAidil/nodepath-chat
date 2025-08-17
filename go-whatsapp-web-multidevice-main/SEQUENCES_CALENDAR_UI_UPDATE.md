# Sequences UI Calendar Update

## Changes Made

### New Calendar-Style UI for Create Sequence Modal

#### 1. **Top Section - Sequence Settings**
- **Sequence Name*** (required)
- **Sequence Description*** (required) 
- **Sequence Tag*** (required - previously "Niche/Category")
- **Min Delay** (seconds) - Global setting for all days
- **Max Delay** (seconds) - Global setting for all days
- **Schedule Time** - Time when messages should be sent

#### 2. **Calendar Grid View**
- Shows Days 1-31 in a calendar-like grid (7 columns)
- Each day is clickable
- Visual indicators:
  - Gray border + "Add" = No message set
  - Green border + "Set" = Message configured
  - Hover effect for better UX

#### 3. **Day Message Popup**
When clicking on any day, a modal opens with:
- **Day X Message** title
- **Message** textarea with WhatsApp formatting support
- **Live Preview** showing formatted message in real-time
- **Image (Optional)** with automatic compression to 350KB
- **Finish** button to save the day's content

### How It Works

1. Click "Create Sequence" button
2. Fill in the top section (name, description, tag, delays, time)
3. Click on any day (1-31) to add a message
4. In the popup:
   - Write your message
   - See live preview with formatting
   - Optionally add an image
   - Click "Finish" to save
5. Configured days show green border and "Set" status
6. Configure as many days as needed
7. Click "Create Sequence" to save the entire sequence

### Technical Implementation

- **Data Structure**: `sequenceDays` object stores all day configurations
- **Live Preview**: Real-time formatting for bold, italic, strikethrough, code
- **Image Compression**: Automatic resize and quality adjustment to stay under 350KB
- **Responsive Grid**: 7 columns on desktop, 4 on tablet, 3 on mobile
- **Visual Feedback**: Hover effects and status indicators

### UI Benefits

1. **Visual Overview**: See all 31 days at a glance
2. **Flexible**: Add messages to any days in any order
3. **Intuitive**: Calendar metaphor is familiar to users
4. **Efficient**: Global delay/time settings apply to all days
5. **Preview**: See exactly how messages will appear on WhatsApp

### API Integration

The system still uses the same `/api/sequences` endpoint with the same data structure:
```json
{
  "name": "Sequence Name",
  "description": "Description",
  "niche": "Tag",
  "steps": [
    {
      "day_number": 1,
      "content": "Message text",
      "image_url": "base64...",
      "min_delay_seconds": 5,
      "max_delay_seconds": 15
    }
  ],
  "status": "draft"
}
```

## Testing

1. Open Dashboard
2. Go to Sequences tab
3. Click "Create Sequence"
4. Fill in details
5. Click various days to add messages
6. Notice the visual feedback
7. Create the sequence
8. Verify it appears in the list
