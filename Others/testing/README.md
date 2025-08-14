# WhatsApp Multi-Device Testing System

This testing system allows you to simulate the entire WhatsApp Multi-Device system with 3000 devices without sending any real WhatsApp messages.

## 🎯 Purpose

- Test system stability with high device count
- Verify campaign and sequence processing
- Monitor performance metrics
- Identify bottlenecks before production
- No real WhatsApp API calls - completely safe

## 📁 Files

1. **generate_test_data.sql** - SQL script to create 3000 test devices and 50,000 leads
2. **mock_whatsapp_client.go** - Mock WhatsApp client that simulates sending
3. **test_runner.go** - Main test application with interactive menu
4. **performance_monitor.html** - Real-time performance dashboard
5. **run_test.bat** - Batch file to run the test system

## 🚀 Quick Start

### 1. Generate Test Data

First, run the SQL script to generate test data:

```bash
# Option 1: Using psql
psql -U postgres -d whatsapp -f generate_test_data.sql

# Option 2: Using the test runner menu
run_test.bat
# Then select option 1
```

This creates:
- 1 test user (test@whatsapp.com / test123)
- 3000 test devices (90% online, 10% offline)
- 50,000 test leads with various triggers
- 3 test campaigns
- 4 test sequences with 30 steps each
- 2 test AI campaigns

### 2. Run Test System

```bash
# Run the interactive test menu
run_test.bat
```

Menu options:
1. **Generate Test Data** - Creates all test data
2. **Simulate Campaign Broadcasting** - Tests campaign sending
3. **Simulate Sequence Processing** - Tests drip campaigns
4. **Simulate AI Campaign** - Tests AI lead distribution
5. **Run Full System Test** - Runs everything
6. **Show Statistics** - Display current metrics
7. **Clean Test Data** - Remove all test data

### 3. Monitor Performance

Open `performance_monitor.html` in your browser to see:
- Real-time message rate graphs
- Device load distribution
- Campaign progress bars
- Success/failure rates
- Top performing devices

## 📊 Test Scenarios

### Campaign Test
- Simulates sending to all matching leads
- Uses multiple devices in parallel
- Applies human-like delays (5-15 seconds)
- 2% simulated failure rate
- Shows messages per second

### Sequence Test
- Tests trigger-based drip campaigns
- Simulates daily message sending
- Verifies lead matching by trigger
- Tests multi-step sequences

### AI Campaign Test
- Simulates intelligent lead distribution
- Tests device load balancing
- Applies throttling rules
- Monitors device limits

## 🔧 Configuration

Edit these values in `test_runner.go`:

```go
simulateDelay: true,           // Enable/disable delays
minDelay:      5 * time.Second,  // Minimum delay
maxDelay:      15 * time.Second, // Maximum delay  
failureRate:   0.02,           // 2% failure rate
workers:       50,             // Parallel workers
```

## 📈 Expected Performance

With 3000 devices:
- **Message Rate**: 150-350 msg/sec (simulated)
- **Hourly Capacity**: 540,000 - 1,260,000 messages
- **Success Rate**: ~98%
- **Device Utilization**: 90% online devices

## 🛡️ Safety Features

- **No Real WhatsApp API Calls**: Everything is simulated
- **Database Isolation**: Uses test user account
- **Easy Cleanup**: One-click removal of all test data
- **Non-Destructive**: Won't affect production data

## 🔍 What to Look For

1. **Stability**: System should handle 3000 devices without crashes
2. **Performance**: Consistent message rates
3. **Memory Usage**: Monitor RAM consumption
4. **Database Load**: Check query performance
5. **Error Handling**: 2% failures should be handled gracefully

## 🚨 Troubleshooting

### Database Connection Error
```
Error: Failed to connect to database
```
Solution: Set DATABASE_URL environment variable or update connection string

### Build Error
```
Error: Build failed
```
Solution: Ensure Go is installed and in PATH

### No Devices Found
```
Error: No online devices found
```
Solution: Run option 1 to generate test data first

## 📝 Test Results Interpretation

### Good Performance Indicators:
- Message rate > 100 msg/sec
- Success rate > 95%
- All campaigns completing
- No memory leaks
- Stable device connections

### Warning Signs:
- Message rate < 50 msg/sec
- Success rate < 90%
- Increasing memory usage
- Database connection errors
- Device disconnections

## 🎯 Next Steps

After successful testing:

1. **Analyze Results**: Review performance metrics
2. **Optimize**: Address any bottlenecks found
3. **Scale Testing**: Try with different loads
4. **Production Ready**: Deploy with confidence

## ⚠️ Important Notes

- This is a **simulation only** - no real messages are sent
- Test data uses "TestDevice" and "TestLead" prefixes
- Always clean up test data after testing
- Don't run on production database without backup

## 🤝 Integration with Main System

To use mock clients in the main system for testing:

1. Set environment variable: `MOCK_MODE=true`
2. The system will use mock clients instead of real WhatsApp
3. All operations will be simulated
4. Check logs for `[MOCK]` prefix

Happy Testing! 🚀
