package main

import (
    "fmt"
)

// FIX FOR SEQUENCE LEADS CALCULATION DISCREPANCY
// ==============================================

// Problem: Total leads are being summed across steps instead of counting unique leads

// BACKEND FIX 1: Fix the sequence summary total leads calculation
// Location: app.go around line 2360
func fixSequenceSummaryTotalLeads() {
    fmt.Println(`
REPLACE THIS CODE in app.go GetSequenceSummary():

    // Calculate totals from individual sequences
    for _, seq := range sequencesWithFlows {
        if leads, ok := seq["total_leads"].(int); ok {
            totalLeadsSum += leads  // This SUMS leads from each sequence
        }
    }

WITH THIS CODE:

    // Get UNIQUE total leads across all sequences (not sum)
    if db != nil {
        var uniqueTotalLeads int
        
        // Build query with optional date filter
        query := ` + "`" + `
            SELECT COUNT(DISTINCT CONCAT(recipient_phone, '|', device_id)) 
            FROM broadcast_messages
            WHERE user_id = ?
            AND sequence_id IS NOT NULL` + "`" + `
        
        args := []interface{}{session.UserID}
        
        if startDate != "" && endDate != "" {
            query += ` + "`" + ` AND DATE(scheduled_at) BETWEEN ? AND ?` + "`" + `
            args = append(args, startDate, endDate)
        } else if startDate != "" {
            query += ` + "`" + ` AND DATE(scheduled_at) >= ?` + "`" + `
            args = append(args, startDate)
        } else if endDate != "" {
            query += ` + "`" + ` AND DATE(scheduled_at) <= ?` + "`" + `
            args = append(args, endDate)
        }
        
        err := db.QueryRow(query, args...).Scan(&uniqueTotalLeads)
        if err == nil {
            totalLeadsSum = uniqueTotalLeads
        } else {
            log.Printf("Error getting unique total leads: %v", err)
            totalLeadsSum = 0
        }
    }
`)
}

// FRONTEND FIX: Add clarification to step statistics display
// Location: dashboard.html around line 6900
func fixFrontendStepStatisticsDisplay() {
    fmt.Println(`
ADD THIS NOTE in dashboard.html after the step statistics cards:

    </div>
    <!-- Add clarification note -->
    <div class="row mt-3">
        <div class="col-12">
            <div class="alert alert-info">
                <i class="fas fa-info-circle"></i> 
                <strong>Note:</strong> The "Total Leads" shown in each step may include the same contacts 
                who progress through multiple steps. The overall "Total Leads" count above shows the 
                unique number of contacts across all steps.
            </div>
        </div>
    </div>
`)
}

// Alternative: Change the label to be clearer
func alternativeFixStepLabels() {
    fmt.Println(`
ALTERNATIVE FIX - Change the label in step statistics cards:

REPLACE:
    <small>Total Leads</small>

WITH:
    <small>Step Recipients</small>
`)
}

func main() {
    fmt.Println("SEQUENCE LEADS CALCULATION FIX")
    fmt.Println("==============================\n")
    
    fmt.Println("The issue: Step-wise 'Total Leads' are being summed up, but they overlap!")
    fmt.Println("Example: If John receives 3 steps, he's counted 3 times instead of 1\n")
    
    fixSequenceSummaryTotalLeads()
    fmt.Println("\n" + "="*60 + "\n")
    
    fixFrontendStepStatisticsDisplay()
    fmt.Println("\n" + "="*60 + "\n")
    
    alternativeFixStepLabels()
    
    fmt.Println("\n\nIMPLEMENTATION STEPS:")
    fmt.Println("1. Apply the backend fix to get correct unique total leads")
    fmt.Println("2. Add the clarification note to the frontend")
    fmt.Println("3. Optionally change 'Total Leads' to 'Step Recipients' for clarity")
    fmt.Println("4. Test with a sequence that has multiple steps to verify the fix")
}
