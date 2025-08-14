# Test Webhook Script

## 1. First, set the webhook key in your Railway environment variables:
```
WEBHOOK_LEAD_KEY=my-super-secret-key-2025
```

## 2. Get your User ID:
- Login to your WhatsApp Multi-Device admin dashboard
- Go to Settings or Profile
- Copy your User ID (UUID format)

## 3. Test the webhook with curl:

### Basic Lead Creation (Prospect)
```bash
curl -X POST https://your-app.railway.app/webhook/lead/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Lead from Bot",
    "phone": "60123456789",
    "target_status": "prospect",
    "user_id": "YOUR-USER-UUID-HERE",
    "webhook_key": "my-super-secret-key-2025"
  }'
```

### Full Lead Creation with All Fields
```bash
curl -X POST https://your-app.railway.app/webhook/lead/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "phone": "60123456789",
    "email": "john@example.com",
    "niche": "fitness",
    "source": "whatsapp_bot_campaign_jan2025",
    "target_status": "customer",
    "trigger": "fitness_start,welcome_sequence",
    "notes": "Interested in weight loss program, contacted via WhatsApp bot",
    "user_id": "YOUR-USER-UUID-HERE",
    "device_id": "OPTIONAL-DEVICE-UUID",
    "webhook_key": "my-super-secret-key-2025"
  }'
```

## 4. Node.js Example for Your WhatsApp Bot:

```javascript
const axios = require('axios');

// Configuration
const WEBHOOK_URL = 'https://your-app.railway.app/webhook/lead/create';
const WEBHOOK_KEY = 'my-super-secret-key-2025';
const USER_ID = 'YOUR-USER-UUID-HERE';

// Function to create lead when bot receives new contact
async function createLeadFromWhatsApp(contactInfo) {
  try {
    const leadData = {
      name: contactInfo.name || contactInfo.pushName || 'Unknown',
      phone: contactInfo.number.replace(/[^0-9]/g, ''), // Remove non-numeric chars
      target_status: 'prospect', // New contacts are prospects
      source: 'whatsapp_bot',
      trigger: determineSequenceTrigger(contactInfo.firstMessage),
      notes: `First message: ${contactInfo.firstMessage}`,
      user_id: USER_ID,
      webhook_key: WEBHOOK_KEY
    };

    const response = await axios.post(WEBHOOK_URL, leadData);
    console.log('Lead created:', response.data);
    return response.data.results.lead_id;
  } catch (error) {
    console.error('Failed to create lead:', error.response?.data || error.message);
    return null;
  }
}

// Helper function to determine sequence trigger based on message
function determineSequenceTrigger(message) {
  const lowerMessage = message.toLowerCase();
  
  if (lowerMessage.includes('fitness') || lowerMessage.includes('gym')) {
    return 'fitness_start';
  } else if (lowerMessage.includes('crypto') || lowerMessage.includes('bitcoin')) {
    return 'crypto_welcome';
  } else if (lowerMessage.includes('business') || lowerMessage.includes('entrepreneur')) {
    return 'business_inquiry';
  }
  
  return 'general_inquiry';
}

// Example usage in your bot
bot.on('message', async (message) => {
  // Check if it's a new contact
  if (isNewContact(message.from)) {
    const contactInfo = {
      number: message.from,
      name: message.pushName,
      firstMessage: message.body
    };
    
    const leadId = await createLeadFromWhatsApp(contactInfo);
    if (leadId) {
      // Lead created successfully, continue with bot flow
      await bot.sendMessage(message.from, 'Welcome! How can I help you today?');
    }
  }
});
```

## 5. Python Example:

```python
import requests
import re

# Configuration
WEBHOOK_URL = 'https://your-app.railway.app/webhook/lead/create'
WEBHOOK_KEY = 'my-super-secret-key-2025'
USER_ID = 'YOUR-USER-UUID-HERE'

def create_lead_from_whatsapp(contact_info):
    """Create a lead when bot receives new contact"""
    
    # Clean phone number (remove non-numeric characters)
    phone = re.sub(r'[^0-9]', '', contact_info['number'])
    
    lead_data = {
        'name': contact_info.get('name', 'Unknown'),
        'phone': phone,
        'target_status': 'prospect',
        'source': 'whatsapp_bot',
        'trigger': determine_sequence_trigger(contact_info.get('first_message', '')),
        'notes': f"First message: {contact_info.get('first_message', 'N/A')}",
        'user_id': USER_ID,
        'webhook_key': WEBHOOK_KEY
    }
    
    try:
        response = requests.post(WEBHOOK_URL, json=lead_data)
        response.raise_for_status()
        
        result = response.json()
        print(f"Lead created: {result['results']['lead_id']}")
        return result['results']['lead_id']
    except requests.exceptions.RequestException as e:
        print(f"Failed to create lead: {e}")
        if hasattr(e.response, 'json'):
            print(f"Error details: {e.response.json()}")
        return None

def determine_sequence_trigger(message):
    """Determine sequence trigger based on message content"""
    message_lower = message.lower()
    
    if 'fitness' in message_lower or 'gym' in message_lower:
        return 'fitness_start'
    elif 'crypto' in message_lower or 'bitcoin' in message_lower:
        return 'crypto_welcome'
    elif 'business' in message_lower or 'entrepreneur' in message_lower:
        return 'business_inquiry'
    
    return 'general_inquiry'
```

## 6. Testing Locally (before deploying to Railway):

If you want to test locally first:

1. Run the WhatsApp Multi-Device app locally:
```bash
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
whatsapp.exe rest
```

2. Test with local URL:
```bash
curl -X POST http://localhost:3000/webhook/lead/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Local Test Lead",
    "phone": "60123456789",
    "target_status": "prospect",
    "user_id": "YOUR-USER-UUID",
    "webhook_key": "your-secret-webhook-key-here"
  }'
```

## 7. Integration Flow:

```
WhatsApp User → Your WhatsApp Bot → Webhook API → Lead Created in DB
                                                 ↓
                                          Sequences Triggered
                                                 ↓
                                          Campaigns Can Target
```

## 8. Important Notes:

- Phone numbers should be without spaces or special characters
- The `trigger` field accepts comma-separated values for multiple sequences
- Leads with `target_status: "customer"` won't be included in prospect campaigns
- The webhook is rate-limited by Railway's default settings
- All timestamps are in UTC
