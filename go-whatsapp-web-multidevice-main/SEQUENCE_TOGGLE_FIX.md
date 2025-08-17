# Fix Summary: Sequence Toggle Active/Inactive

## Problem
When toggling a sequence to inactive, it wasn't updating the status properly. The UI would still show "active" and the badge color wouldn't change to red.

## Root Cause
The `GetSequenceByID` and `GetSequences` queries in the repository were NOT selecting the `status` column from the database. This caused the toggle function to always read the status as empty/null, making the toggle logic fail.

## Solution Applied

### 1. Fixed Repository Queries (src/repository/sequence_repository.go)
- Added `status` to the SELECT query in `GetSequenceByID` function
- Added `status` to the SELECT query in `GetSequences` function
- Added `&seq.Status` to the Scan parameters to properly read the status value

### 2. Updated Frontend Colors (src/views/dashboard.html)
- Added color mapping for 'inactive' status to show 'danger' (red) color
- Ensured the badge and progress bar properly reflect the inactive state

### 3. CSS Enhancements
- Added styles for toggle switches to show green when active and red when inactive
- Added hover effects for sequence cards

## How It Works Now
1. When you toggle a sequence, it properly reads the current status from the database
2. The backend toggles between 'active' and 'inactive' states
3. The frontend reloads the sequences list and displays the correct status with appropriate colors:
   - Active = Green badge
   - Inactive = Red badge
   - Paused = Yellow badge

## Testing
1. Build and run the application
2. Go to the Sequences tab in the dashboard
3. Toggle any sequence - it should now properly switch between active (green) and inactive (red)
4. The status should persist after page refresh

No database migration needed since you already have the status column!
