# Test script to verify prospect_name updating functionality
# This script sends multiple webhooks with different names to test if prospect_name gets updated

$baseUrl = "http://localhost:8080"
$webhookEndpoint = "$baseUrl/api/ai-whatsapp/webhook/waha/FakhriAidilTLW-001"

# Test 1: Send webhook with first name
Write-Host "Test 1: Sending webhook with name 'TestUser1'..." -ForegroundColor Yellow
$payload1 = @{
    event = "message"
    session = "FakhriAidilTLW-001"
    payload = @{
        _data = @{
            id = @{
                fromMe = $false
                remote = "601137508067@c.us"
                id = "test_message_1"
            }
            body = "Hello, this is test message 1"
            type = "chat"
            timestamp = [int][double]::Parse((Get-Date -UFormat %s))
            from = "601137508067@c.us"
            to = "FakhriAidilTLW-001@c.us"
            fromMe = $false
            hasMedia = $false
            ack = 1
        }
        media = @{
            Info = @{
                PushName = "TestUser1"
            }
        }
    }
} | ConvertTo-Json -Depth 10

try {
    $response1 = Invoke-RestMethod -Uri $webhookEndpoint -Method POST -Body $payload1 -ContentType "application/json"
    Write-Host "Response 1: $($response1 | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "Error in Test 1: $($_.Exception.Message)" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# Test 2: Send webhook with different name to same device/phone
Write-Host "Test 2: Sending webhook with name 'TestUser2' to same device/phone..." -ForegroundColor Yellow
$payload2 = @{
    event = "message"
    session = "FakhriAidilTLW-001"
    payload = @{
        _data = @{
            id = @{
                fromMe = $false
                remote = "601137508067@c.us"
                id = "test_message_2"
            }
            body = "Hello, this is test message 2 with updated name"
            type = "chat"
            timestamp = [int][double]::Parse((Get-Date -UFormat %s))
            from = "601137508067@c.us"
            to = "FakhriAidilTLW-001@c.us"
            fromMe = $false
            hasMedia = $false
            ack = 1
        }
        media = @{
            Info = @{
                PushName = "TestUser2"
            }
        }
    }
} | ConvertTo-Json -Depth 10

try {
    $response2 = Invoke-RestMethod -Uri $webhookEndpoint -Method POST -Body $payload2 -ContentType "application/json"
    Write-Host "Response 2: $($response2 | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "Error in Test 2: $($_.Exception.Message)" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# Test 3: Send webhook with third name to verify continuous updating
Write-Host "Test 3: Sending webhook with name 'TestUser3' to verify continuous updating..." -ForegroundColor Yellow
$payload3 = @{
    event = "message"
    session = "FakhriAidilTLW-001"
    payload = @{
        _data = @{
            id = @{
                fromMe = $false
                remote = "601137508067@c.us"
                id = "test_message_3"
            }
            body = "Hello, this is test message 3 with another updated name"
            type = "chat"
            timestamp = [int][double]::Parse((Get-Date -UFormat %s))
            from = "601137508067@c.us"
            to = "FakhriAidilTLW-001@c.us"
            fromMe = $false
            hasMedia = $false
            ack = 1
        }
        media = @{
            Info = @{
                PushName = "TestUser3"
            }
        }
    }
} | ConvertTo-Json -Depth 10

try {
    $response3 = Invoke-RestMethod -Uri $webhookEndpoint -Method POST -Body $payload3 -ContentType "application/json"
    Write-Host "Response 3: $($response3 | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "Error in Test 3: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "\nAll tests completed! Check server logs to verify prospect_name updates." -ForegroundColor Cyan
Write-Host "Expected behavior: prospect_name should update from TestUser1 -> TestUser2 -> TestUser3" -ForegroundColor Cyan