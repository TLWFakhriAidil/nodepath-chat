#!/usr/bin/env python3

# Read current README
with open('README.md', 'r', encoding='utf-8') as f:
    content = f.read()

# Find where to insert the new update
latest_update_pos = content.find('## 🚀 LATEST UPDATE:')
if latest_update_pos == -1:
    latest_update_pos = content.find('**Deploy**:') + len('**Deploy**: ✅ Auto-deployment via Railway (Fully optimized)\n')

# Find where the previous update section starts
prev_update_pos = content.find('## 🚀 Previous Update:', latest_update_pos)
if prev_update_pos == -1:
    prev_update_pos = content.find('### ✅ Device Duplicate Prevention')

# Create new update section
new_update = '''
## 🚀 LATEST UPDATE: Fixed Sequence Contact Processing (January 19, 2025)

### ✅ Fixed Sequence Step Activation Logic
Fixed critical issues with sequence contact progression where steps were activated out of order:
- **Before**: Steps activated by step number (1, 2, 3...) causing issues
- **After**: Steps activated by **earliest `next_trigger_time`** respecting scheduled times
- **Result**: Proper sequence flow without missing or duplicate steps

### 🔧 Technical Fixes Applied

#### **Database Optimizations:**
```sql
-- 1. Unique constraint preventing duplicate active steps
CREATE UNIQUE INDEX idx_one_active_per_contact
ON sequence_contacts(sequence_id, contact_phone)
WHERE status = 'active';

-- 2. Performance index for finding pending steps by time
CREATE INDEX idx_pending_steps_by_time
ON sequence_contacts(sequence_id, contact_phone, next_trigger_time)
WHERE status = 'pending';
```

#### **Query Changes:**
```sql
-- OLD: Activated by step number
WHERE current_step = $3 AND status = 'pending'

-- NEW: Activated by earliest scheduled time
WHERE status = 'pending' AND next_trigger_time <= NOW()
ORDER BY next_trigger_time ASC
FOR UPDATE SKIP LOCKED  -- Handles 3000 concurrent devices
```

### 📊 What This Fixes:
1. **Missing Step 1**: Steps are now marked 'completed' not deleted
2. **Duplicate Step 2**: Unique constraint prevents multiple active steps
3. **Out-of-order Steps**: Activation by time ensures proper sequence
4. **Race Conditions**: FOR UPDATE SKIP LOCKED prevents conflicts

### 🚀 3000 Device Optimization:
- **Concurrent Processing**: Each device can process without blocking others
- **Row-level Locking**: Prevents duplicate processing of same contact
- **Transaction Isolation**: Proper isolation level for concurrent access
- **Performance Indexes**: Fast lookups even with millions of records

## 🚀 Previous Update: Enhanced Duplicate Prevention (January 19, 2025)
'''

# Update the header
content = content.replace(
    'Last Updated: January 19, 2025 - Enhanced Duplicate Prevention & Database Management',
    'Last Updated: January 19, 2025 - Sequence Processing Fix for 3000 Devices'
)

# Replace the LATEST UPDATE section
if '## 🚀 LATEST UPDATE:' in content:
    # Replace existing latest update
    new_content = content[:latest_update_pos] + new_update + content[prev_update_pos:]
else:
    # Insert new update
    new_content = content[:latest_update_pos] + '\n' + new_update + '\n' + content[latest_update_pos:]

# Write updated README
with open('README.md', 'w', encoding='utf-8') as f:
    f.write(new_content)

print("README.md updated successfully!")