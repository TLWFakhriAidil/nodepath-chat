# AI WhatsApp NodePath Table Column Analysis

## Schema Columns (from database)
Based on the current `ai_whatsapp_nodepath` table schema:

1. `id` - Primary key
2. `id_prospect` - Foreign key to prospects
3. `id_device` - Device identifier
4. `prospect_num` - Phone number
5. `date_order` - Order date
6. `niche` - Business niche
7. `intro` - Introduction text
8. `conv_last` - Last conversation value
9. `conv_current` - Current conversation value
10. `conv_stage` - Conversation stage
11. `variables` - Variables storage
12. `stage` - Current stage
13. `current_node` - Current node in flow
14. `execution_status` - Flow execution status
15. `bot_balas` - Bot reply timestamp
16. `balas` - Reply count/flag
17. `jam` - Time/hour
18. `human` - Human takeover flag
19. `keywordiklan` - Advertisement keyword
20. `catatan_staff` - Staff notes
21. `data_image` - Image data
22. `marketer` - Marketer identifier
23. `created_at` - Creation timestamp
24. `updated_at` - Update timestamp
25. `update_today` - Today's update flag
26. `waiting_for_reply` - Waiting for user reply flag
27. `flow_id` - Flow identifier
28. `last_node_id` - Last node identifier

## Used Columns (from code analysis)
Based on INSERT and UPDATE operations found in the codebase:

### INSERT Operations Use:
- `id_prospect` (IDProspect)
- `id_device` (IDDevice)
- `prospect_num` (ProspectNum)
- `stage` (Stage)
- `date_order` (DateOrder)
- `conv_last` (convLastValue)
- `conv_current` (convCurrentValue)
- `human` (Human)
- `niche` (Niche)
- `jam` (Jam)
- `intro` (Intro)
- `catatan_staff` (CatatanStaff)
- `balas` (Balas)
- `data_image` (DataImage)
- `conv_stage` (ConvStage)
- `bot_balas` (BotBalas)
- `keywordiklan` (KeywordIklan)
- `marketer` (Marketer)
- `update_today` (UpdateToday)
- `created_at` (CreatedAt)
- `updated_at` (UpdatedAt)

### UPDATE Operations Use:
- `id_device`
- `stage`
- `date_order`
- `conv_last`
- `conv_current`
- `human`
- `niche`
- `jam`
- `intro`
- `catatan_staff`
- `balas`
- `data_image`
- `conv_stage`
- `bot_balas`
- `keywordiklan`
- `marketer`
- `update_today`
- `current_node_id` (flow-related)
- `waiting_for_reply` (flow-related)
- `flow_id` (flow-related)
- `last_node_id` (flow-related)
- `updated_at`

### SELECT Operations Use:
- `id_prospect`
- `id_device`
- `prospect_num`
- `stage`
- `date_order`
- `conv_last`
- `conv_current`
- `human`
- `niche`
- `jam`
- `intro`
- `catatan_staff`
- `balas`
- `data_image`
- `conv_stage`
- `bot_balas`
- `keywordiklan`
- `marketer`
- `update_today`
- `created_at`
- `updated_at`

## UNUSED COLUMNS IDENTIFIED
Columns that exist in schema but are NEVER used in any INSERT, UPDATE, or SELECT operations:

1. **`variables`** - Text field for variables storage (never written to or read)
2. **`current_node`** - Old current node field (replaced by `current_node_id`)
3. **`execution_status`** - Flow execution status (never used in any operations)

## ANALYSIS NOTES

### Flow-Related Columns
The flow system uses these columns actively:
- `current_node_id` (actively used in database operations)
- `waiting_for_reply` (actively used)
- `flow_id` (actively used)
- `last_node_id` (actively used)

### Core Conversation Columns
All core conversation columns are actively used:
- `stage`, `conv_last`, `conv_current`, `conv_stage`
- `human`, `balas`, `bot_balas`
- User data: `prospect_num`, `niche`, `jam`, `intro`
- Staff data: `catatan_staff`, `marketer`, `keywordiklan`
- Media: `data_image`
- Timestamps: `created_at`, `updated_at`, `update_today`

### Column Evolution
Based on migration history:
- `current_node` was added in migration 000015
- `current_node_id` was added in migration 000017
- Only `current_node_id` is used in actual database operations
- `current_node` appears to be legacy/unused

## FINAL RECOMMENDATION
Safely remove these unused columns:
1. **`variables`** - Never used in any database operations
2. **`current_node`** - Legacy column, replaced by `current_node_id`
3. **`execution_status`** - Never used in any database operations

These columns can be safely removed as they don't impact the node flow logic functionality.