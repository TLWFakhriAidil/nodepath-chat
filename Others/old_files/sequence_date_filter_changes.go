package main

// This shows all the changes needed for sequence summary date filters

// 1. Update all showTodayOnly references to use date range filters
// Replace all occurrences of:
//   if showTodayOnly {
//       query += ` AND DATE(scheduled_at) = CURDATE()`
//   }
// With:
//   if startDate != "" && endDate != "" {
//       query += ` AND DATE(scheduled_at) BETWEEN ? AND ?`
//       args = append(args, startDate, endDate)
//   } else if startDate != "" {
//       query += ` AND DATE(scheduled_at) >= ?`
//       args = append(args, startDate)  
//   } else if endDate != "" {
//       query += ` AND DATE(scheduled_at) <= ?`
//       args = append(args, endDate)
//   }

// 2. Update the per-sequence query (around line 2144):
//   if showTodayOnly {
//       query += ` AND DATE(scheduled_at) = CURDATE()`
//   }

// 3. Update the overall statistics query (around line 2190):
//   if showTodayOnly {
//       query += ` AND DATE(scheduled_at) = CURDATE()`
//   }

// 4. Pass the date filter to device report
// When calling showSequenceDeviceReport, pass the date filter parameters
