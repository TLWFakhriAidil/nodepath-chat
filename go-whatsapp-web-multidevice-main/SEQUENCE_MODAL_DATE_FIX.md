# Sequence Modal Date Filter Fix Instructions

## Issue
When clicking on the success/failed count in the sequence summary, the modal shows ALL messages ever sent for that step, ignoring the selected date filter.

## Files to Modify

### 1. `src/ui/rest/app.go` (Backend)

Find the `GetSequenceStepLeads` function (around line 4906).

**Step 1:** After this line (around line 4911):
```go
status := c.Query("status", "all")
```

Add:
```go
// Get date filters from query params
startDate := c.Query("start_date")
endDate := c.Query("end_date")
```

**Step 2:** After this line (around line 4947):
```go
args := []interface{}{sequenceId, deviceId, stepId, session.UserID}
```

Add:
```go
log.Printf("GetSequenceStepLeads - Sequence: %s, Device: %s, Step: %s, Status: %s, DateRange: %s to %s",
    sequenceId, deviceId, stepId, status, startDate, endDate)
```

**Step 3:** After the status filter conditions, before `query += ORDER BY bm.sent_at DESC` (around line 4959):

Add:
```go
// Add date filter if provided
if startDate != "" && endDate != "" {
    query += ` AND DATE(bm.sent_at) BETWEEN ? AND ?`
    args = append(args, startDate, endDate)
} else if startDate != "" {
    query += ` AND DATE(bm.sent_at) >= ?`
    args = append(args, startDate)
} else if endDate != "" {
    query += ` AND DATE(bm.sent_at) <= ?`
    args = append(args, endDate)
}
```

### 2. `src/views/dashboard.html` (Frontend)

Find the `showSequenceStepLeadDetails` function (around line 7358).

After this line:
```javascript
let url = `/api/sequences/${currentSequenceForReport.id}/device/${deviceId}/step/${stepId}/leads?status=${status}`;
```

Add:
```javascript
// Get the current date filters from the sequence summary
const startDate = document.getElementById('sequenceStartDate').value;
const endDate = document.getElementById('sequenceEndDate').value;

if (startDate) {
    url += `&start_date=${startDate}`;
}
if (endDate) {
    url += `&end_date=${endDate}`;
}

console.log('Fetching sequence step leads with URL:', url);
```

## Testing

1. Go to Sequences tab
2. Set a date filter (e.g., Today)
3. Click on a success/failed count
4. Verify that the modal only shows messages from the selected date range

## Build and Deploy

```bash
# Build without CGO
set CGO_ENABLED=0
go build

# Commit and push
git add -A
git commit -m "Fix: Add date filter to sequence step leads modal"
git push origin main
```
