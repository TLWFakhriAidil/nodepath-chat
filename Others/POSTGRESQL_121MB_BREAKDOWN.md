# What's in the 121 MB PostgreSQL Database?

Based on the analysis, here's the breakdown of the 121 MB PostgreSQL database:

## Storage Breakdown by Category:

### 1. **WhatsApp Session Data (73.9 MB - 61.1%)**
These are `whatsmeow_*` tables that store WhatsApp Web session data:

| Table | Size | Rows | Purpose |
|-------|------|------|---------|
| whatsmeow_message_secrets | 37 MB | 120,834 | End-to-end encryption keys for messages |
| whatsmeow_contacts | 18 MB | 98,987 | Contact information and profiles |
| whatsmeow_app_state_mutation_macs | 16 MB | 55,215 | App state synchronization |
| whatsmeow_sessions | 904 KB | 689 | Active device sessions |
| whatsmeow_sender_keys | 320 KB | 481 | Group message encryption |
| Other whatsmeow tables | ~2 MB | Various | Device info, settings, etc. |

**Total: 15 tables storing device sessions & encryption data**

### 2. **WhatsApp Chat/Message Data (39 MB - 32.2%)**

| Table | Size | Rows | Purpose |
|-------|------|------|---------|
| whatsapp_messages | 28 MB | 43,511 | Message history (text, media URLs, timestamps) |
| whatsapp_chats | 11 MB | 33,926 | Chat metadata (names, groups, last message) |

### 3. **Application Data (~0.9 MB - 0.7%)**
Small tables for app functionality:
- users (5 users)
- user_devices (87 devices)
- leads (0 - cleared)
- campaigns (0 - cleared)
- sequences (0 - cleared)
- Other system tables

## Key Findings:

1. **93.3% is WhatsApp-related data** (Session + Chat data)
2. **Only 0.7% is application data** (mostly cleared in cleanup)
3. **Largest single table**: whatsmeow_message_secrets (37 MB - 30.6%)

## Why So Much Space for WhatsApp Sessions?

1. **Encryption Keys (37 MB)**: WhatsApp stores encryption keys for every message to maintain end-to-end encryption
2. **Contacts (18 MB)**: Stores full contact details for ~99,000 contacts
3. **App State (16 MB)**: Synchronization data for WhatsApp Web
4. **Message History (28 MB)**: 43,511 messages with content and metadata
5. **Chat Data (11 MB)**: Information about 33,926 chats

## Recommendations to Reduce Further:

### Safe to Clean (Won't affect active devices):
1. **Old message encryption keys** (~20 MB potential savings)
   ```sql
   DELETE FROM whatsmeow_message_secrets 
   WHERE jid IN (SELECT jid FROM user_devices WHERE last_seen < NOW() - INTERVAL '30 days');
   ```

2. **Old messages** (~15 MB potential savings)
   ```sql
   DELETE FROM whatsapp_messages WHERE created_at < NOW() - INTERVAL '30 days';
   ```

3. **Duplicate contacts** (~5 MB potential savings)
   ```sql
   -- Remove duplicates keeping most recent
   DELETE FROM whatsmeow_contacts a USING whatsmeow_contacts b
   WHERE a.id < b.id AND a.jid = b.jid;
   ```

### Potential Final Size: 60-70 MB

After aggressive cleanup, the database could be reduced to 60-70 MB, which would be:
- Core WhatsApp session data for active devices
- Recent messages (last 30 days)
- Active contacts only
- Minimal application data

## Summary:
The 121 MB consists primarily of WhatsApp Web session data (encryption keys, contacts, and message history). This is normal for a system managing 87 WhatsApp devices with extensive message history and nearly 100,000 contacts.
