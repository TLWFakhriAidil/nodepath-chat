# Lead Management System Updates

## Changes Made (June 27, 2025)

### 1. Phone Number Format
- **Old**: Placeholder showed "+60123456789"
- **New**: Placeholder shows "60123456789" (without +)
- Phone numbers should be entered without the + symbol

### 2. Niche Field
- **Old**: Label was "Niche/Industry" with placeholder "e.g., Real Estate, E-commerce"
- **New**: Label is just "Niche" with placeholder "EXSTART or EXSTART,ITADRESS"
- Supports single niche: `EXSTART`
- Supports multiple niches: `EXSTART,ITADRESS` (comma-separated)

### 3. Status Options
- **Old**: new, contacted, qualified, converted, lost
- **New**: 
  - `prospect` - Potential customer
  - `customer` - Converted customer

### 4. Journey/Notes Field
- **Old**: Label was "Journey/Notes"
- **New**: Label is "Additional Note"
- Placeholder updated to "Enter additional notes about this lead..."

### 5. Filter System
- **Status Filter**: Shows All, Prospect, Customer
- **Niche Filter**: Dynamically builds from unique niches in the database
  - Shows "All" by default
  - Automatically adds filter buttons for each unique niche found
  - Supports filtering by individual niches even if leads have multiple niches

### 6. Import/Export Updates
- **Import CSV Format**:
  - name (required)
  - phone (required - format: 60123456789)
  - niche (optional - can be single: EXSTART or multiple: EXSTART,ITADRESS)
  - additional_note (optional)
  - status (optional: prospect, customer)

- **Export CSV Headers**: Updated to use "additional_note" instead of "journey"

### 7. Default Values
- New leads default to status: `prospect`
- Empty niche field is allowed
- Additional note is optional

## Usage Examples

### Adding a Lead
1. Click "Add Lead" button
2. Enter name: "Aqil"
3. Enter phone: "60123456789" (no + symbol)
4. Enter niche: "EXSTART" or "EXSTART,ITADRESS"
5. Select status: "Prospect" or "Customer"
6. Enter additional note (optional)
7. Click "Save Lead"

### Filtering Leads
1. **By Status**: Click on "Prospect" or "Customer" filter chips
2. **By Niche**: Click on any niche filter that appears (e.g., "EXSTART", "ITADRESS")
3. **By Search**: Type in the search box to search by name, phone, or niche

### Importing Leads
Create a CSV file with this format:
```csv
name,phone,niche,status,additional_note
Aqil,60123456789,EXSTART,prospect,First contact from website
John,60987654321,"EXSTART,ITADRESS",customer,Purchased both services
```

### API Changes
The backend API remains compatible. The lead structure supports:
- Any status value (not limited to specific options)
- Niche field as a string (frontend handles comma-separated values)
- Journey field (shown as "Additional Note" in UI)
