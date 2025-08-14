# Campaign Summary Update - Implementation Summary

## Changes Made:

### 1. Backend Updates (Go)

#### Campaign Repository (`src/repository/campaign_repository.go`)
- Added two new methods to the `CampaignRepository` interface:
  - `GetCampaignBroadcastStats(campaignID int) (shouldSend, doneSend, failedSend int, err error)`
  - `GetUserCampaignBroadcastStats(userID string) (shouldSend, doneSend, failedSend int, err error)`

- Implemented these methods to calculate:
  - **shouldSend**: Total leads matching campaign's niche and target_status
  - **doneSend**: Count of broadcast_messages with status='sent'
  - **failedSend**: Count of broadcast_messages with status='failed'
  - **remainingSend**: Calculated as shouldSend - doneSend - failedSend

#### App Handler (`src/ui/rest/app.go`)
- Updated `GetCampaignSummary` function to:
  - Call the new repository methods to get broadcast statistics
  - Add `broadcast_stats` object to the response with:
    - `total_should_send`
    - `total_done_send`
    - `total_failed_send`
    - `total_remaining_send`
  - Enhanced `recent_campaigns` array to include per-campaign statistics

### 2. Frontend Updates (HTML/JavaScript)

#### Dashboard HTML (`src/views/dashboard.html`)
- Removed the "Message Statistics" section as requested
- Updated `displayCampaignSummary` function to:
  - Display new statistic cards matching sequence summary format
  - Added 4 new columns to the Recent Campaigns table:
    - Contacts Should Send
    - Contacts Done Send Message
    - Contacts Failed Send Message
    - Contacts Remaining Send Message

## How It Works:

1. **Should Send Calculation**: 
   - Queries the `leads` table
   - Filters by campaign's `user_id`, `niche`, and `target_status`
   - Counts distinct phone numbers

2. **Done/Failed Send Calculation**:
   - Queries the `broadcast_messages` table
   - Filters by `campaign_id`
   - Counts messages by status ('sent' or 'failed')

3. **Remaining Send Calculation**:
   - Simple math: shouldSend - doneSend - failedSend
   - Never shows negative numbers (defaults to 0)

## Visual Changes:

### Before:
- Total Campaigns card
- Status flow cards (Pending, Triggered, Processing, Finished, Failed)
- Message Statistics section with total/sent/failed/success rate
- Basic campaign table

### After:
- Total Campaigns card
- Total Contacts Should Send card
- Contacts Done Send Message card (green)
- Contacts Failed Send Message card (red)
- Contacts Remaining Send Message card (yellow)
- Pending campaigns card
- Enhanced campaign table with 4 new statistic columns

The campaign summary now matches the sequence summary format exactly, providing consistent user experience across both features.
