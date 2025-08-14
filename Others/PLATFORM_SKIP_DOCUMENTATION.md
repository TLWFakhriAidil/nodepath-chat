# Platform Skip Feature Documentation

## Overview
This feature allows you to skip automatic device status checking for devices that have a value in the `platform` column of the `user_devices` table. This is useful for devices that are managed by external systems or have special handling requirements.

## How It Works

When a device has any non-empty value in the `platform` column:
1. **Status Normalizer** (5-minute check) - Skips the device
2. **Auto Connection Monitor** (15-minute check) - Skips the device
3. **Campaign Processing** - Skips the device
4. **Sequence Processing** - Skips the device

## Database Changes

### Add Platform Column
```sql
ALTER TABLE user_devices 
ADD COLUMN IF NOT EXISTS platform VARCHAR(50);

CREATE INDEX IF NOT EXISTS idx_user_devices_platform 
ON user_devices(platform) 
WHERE platform IS NOT NULL AND platform != '';
```

### Example Usage
```sql
-- Set platform for a device to skip it from checks
UPDATE user_devices 
SET platform = 'external_api' 
WHERE id = 'device-uuid-here';

-- Remove platform to resume normal checking
UPDATE user_devices 
SET platform = NULL 
WHERE id = 'device-uuid-here';

-- View all devices with platform set
SELECT id, device_name, status, platform 
FROM user_devices 
WHERE platform IS NOT NULL AND platform != '';
```

## Modified Components

### 1. Device Status Normalizer
- **File**: `infrastructure/whatsapp/device_status_normalizer.go`
- **Change**: Query now includes `WHERE (platform IS NULL OR platform = '')`
- **Effect**: Devices with platform values are not normalized to online/offline

### 2. Auto Connection Monitor
- **File**: `infrastructure/whatsapp/auto_connection_monitor_15min.go`
- **Change**: Query filters out devices with platform values
- **Effect**: No reconnection attempts for platform devices

### 3. Campaign Processor
- **File**: `usecase/optimized_campaign_trigger.go`
- **Change**: Checks platform value before processing device
- **Effect**: Campaigns skip devices with platform set

### 4. Sequence Processor
- **File**: `usecase/sequence_trigger_processor.go`
- **Function**: `getDeviceWorkloads()`
- **Change**: Query includes `AND (d.platform IS NULL OR d.platform = '')`
- **Effect**: Sequences won't use platform devices for sending

## Logging

The system will log when devices are skipped:
```
Device status normalizer: "Status normalization complete: 100 total devices, 10 with platform (skipped)"
Auto connection monitor: "15-minute check complete: 90 devices checked (skipped 10 with platform)"
Campaign processor: "Skipping device Sales Team - has platform: external_api"
```

## Use Cases

1. **External API Devices**: Set `platform = 'external_api'` for devices managed by external systems
2. **Testing Devices**: Set `platform = 'test'` for devices used only for testing
3. **Backup Devices**: Set `platform = 'backup'` for devices kept as backups
4. **Special Handlers**: Set `platform = 'custom_handler'` for devices with special processing

## Implementation Steps

1. **Run SQL Migration**:
   ```bash
   psql -U your_user -d your_database -f add_platform_column.sql
   ```

2. **Apply Code Changes**:
   ```bash
   # Run the batch file to apply changes
   apply_platform_skip.bat
   ```

3. **Build Application**:
   ```bash
   cd src
   go build -o ../whatsapp.exe .
   ```

4. **Test**:
   - Set platform for a test device
   - Monitor logs to confirm device is skipped
   - Verify campaigns/sequences don't use the device

## Rollback

To rollback this feature:
1. Restore original files from `backups/platform_skip/`
2. Remove platform column: `ALTER TABLE user_devices DROP COLUMN platform;`
3. Rebuild application

## Notes

- Platform can be any string value - as long as it's not NULL or empty, the device will be skipped
- This doesn't affect manual operations (refresh, logout, etc.) - only automatic checks
- Devices with platform set can still be used if explicitly specified in campaigns
- The platform column is indexed for performance when many devices have it set
