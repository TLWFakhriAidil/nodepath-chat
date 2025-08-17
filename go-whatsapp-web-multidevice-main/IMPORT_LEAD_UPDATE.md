# Import Lead Update - January 2025

## Changes Made

### 1. Updated Import Modal UI
The import modal now clearly shows only the 5 accepted columns:

- **name** (required) - Lead's name
- **phone** (required) - Phone number in format like 60123456789
- **niche** (required) - Can be single (EXSTART) or multiple (EXSTART,ITADRESS)
- **target_status** (required) - Either "prospect" or "customer"
- **trigger** (optional) - Can be single (EXSTART) or multiple (EXSTART,ITADRESS)

Added an info box that explicitly states: "Only these 5 columns are accepted. Any other columns will be ignored."

### 2. Frontend Validation
- Added validation to ensure all required fields (name, phone, niche, target_status) are present
- Shows warning messages for invalid data
- Validates target_status must be either "prospect" or "customer"
- Shows confirmation dialog before importing

### 3. Export Function Updated
- Export now only includes the 5 columns: name, phone, niche, target_status, trigger
- Removed additional_note and device_id from export

### 4. Backend Validation
- Updated ImportLeads function to require niche column
- Added validation to skip rows with missing required fields
- Removed processing of additional_note, notes, journey, and device_id columns
- All imported leads now use the current device ID

### 5. Files Modified
- `src/views/device_leads.html` - Updated UI and JavaScript functions
- `src/ui/rest/app.go` - Updated ImportLeads and ExportLeads functions

## CSV Format Example

```csv
name,phone,niche,target_status,trigger
John Doe,60123456789,EXSTART,prospect,fitness_start
Jane Smith,60198765432,"EXSTART,ITADRESS",customer,"fitness_start,crypto_welcome"
Bob Johnson,60112345678,ITADRESS,prospect,
```

## Important Notes
1. The niche field is now REQUIRED - leads without niche will be skipped
2. target_status must be exactly "prospect" or "customer" (case-sensitive)
3. Invalid target_status values will default to "prospect" with a warning
4. The device_id is no longer imported - all leads use the current device
5. Additional notes/journey fields are no longer imported
