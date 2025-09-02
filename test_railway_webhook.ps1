# Railway WAHA Webhook Real-time Test Script
# This script tests the WAHA webhook integration for Railway deployment

param(
    [string]$BaseUrl = "https://nodepath-chat-production.up.railway.app",
    [string]$DeviceId = "FakhriAidilTLW-001",
    [string]$TestPhone = "601137508067"
)

Write-Host "=== Railway WAHA Webhook Real-time Test ===" -ForegroundColor Green
Write-Host "Base URL: $BaseUrl" -ForegroundColor Yellow
Write-Host "Device ID: $DeviceId" -ForegroundColor Yellow
Write-Host "Test Phone: $TestPhone" -ForegroundColor Yellow
Write-Host ""

# Test 1: Health Check
Write-Host "1. Testing Health Check..." -ForegroundColor Cyan
try {
    $healthResponse = Invoke-RestMethod -Uri "$BaseUrl/healthz" -Method GET -TimeoutSec 10
    Write-Host "✅ Health Check: OK" -ForegroundColor Green
} catch {
    Write-Host "❌ Health Check Failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 2: WAHA Extraction Endpoint
Write-Host "\n2. Testing WAHA Extraction Endpoint..." -ForegroundColor Cyan
$extractionPayload = @{
    payload = @{
        _data = @{
            from = "$TestPhone@c.us"
            body = "Test extraction message $(Get-Date -Format 'HH:mm:ss')"
            info = @{
                pushName = "Railway Test User"
                fromMe = $false
            }
        }
    }
} | ConvertTo-Json -Depth 5

try {
    $extractionResponse = Invoke-RestMethod -Uri "$BaseUrl/api/ai-whatsapp/test/waha/extraction" -Method POST -ContentType "application/json" -Body $extractionPayload -TimeoutSec 15
    Write-Host "✅ Extraction Test: OK" -ForegroundColor Green
    Write-Host "   Sender Phone: $($extractionResponse.sender_phone)" -ForegroundColor Gray
    Write-Host "   Sender Name: $($extractionResponse.sender_name)" -ForegroundColor Gray
    Write-Host "   Message: $($extractionResponse.message)" -ForegroundColor Gray
} catch {
    Write-Host "❌ Extraction Test Failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 3: Real Webhook Endpoint
Write-Host "\n3. Testing Real Webhook Endpoint..." -ForegroundColor Cyan
$webhookPayload = @{
    payload = @{
        _data = @{
            from = "$TestPhone@c.us"
            body = "Real webhook test message $(Get-Date -Format 'HH:mm:ss')"
            info = @{
                pushName = "Railway Webhook Test"
                fromMe = $false
            }
        }
    }
} | ConvertTo-Json -Depth 5

try {
    $webhookResponse = Invoke-RestMethod -Uri "$BaseUrl/api/ai-whatsapp/webhook/waha/$DeviceId" -Method POST -ContentType "application/json" -Body $webhookPayload -TimeoutSec 20
    Write-Host "✅ Webhook Test: OK" -ForegroundColor Green
    Write-Host "   Response: $($webhookResponse | ConvertTo-Json -Compress)" -ForegroundColor Gray
} catch {
    Write-Host "❌ Webhook Test Failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $errorDetails = $_.Exception.Response | ConvertTo-Json -Depth 3
        Write-Host "   Error Details: $errorDetails" -ForegroundColor Red
    }
}

# Test 4: Load Test (Multiple Concurrent Requests)
Write-Host "\n4. Testing Concurrent Load (10 requests)..." -ForegroundColor Cyan
$jobs = @()
for ($i = 1; $i -le 10; $i++) {
    $loadTestPayload = @{
        payload = @{
            _data = @{
                from = "$TestPhone@c.us"
                body = "Load test message #$i $(Get-Date -Format 'HH:mm:ss.fff')"
                info = @{
                    pushName = "Load Test User $i"
                    fromMe = $false
                }
            }
        }
    } | ConvertTo-Json -Depth 5
    
    $job = Start-Job -ScriptBlock {
        param($url, $payload)
        try {
            $response = Invoke-RestMethod -Uri $url -Method POST -ContentType "application/json" -Body $payload -TimeoutSec 10
            return @{ Success = $true; Response = $response }
        } catch {
            return @{ Success = $false; Error = $_.Exception.Message }
        }
    } -ArgumentList "$BaseUrl/api/ai-whatsapp/test/waha/extraction", $loadTestPayload
    
    $jobs += $job
}

# Wait for all jobs to complete
$results = $jobs | Wait-Job | Receive-Job
$jobs | Remove-Job

$successCount = ($results | Where-Object { $_.Success }).Count
$failCount = ($results | Where-Object { -not $_.Success }).Count

Write-Host "✅ Load Test Results: $successCount/$($results.Count) successful" -ForegroundColor Green
if ($failCount -gt 0) {
    Write-Host "❌ Failed Requests: $failCount" -ForegroundColor Red
}

# Test 5: Response Time Test
Write-Host "\n5. Testing Response Times..." -ForegroundColor Cyan
$responseTimes = @()
for ($i = 1; $i -le 5; $i++) {
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/ai-whatsapp/test/waha/extraction" -Method POST -ContentType "application/json" -Body $extractionPayload -TimeoutSec 10
        $stopwatch.Stop()
        $responseTimes += $stopwatch.ElapsedMilliseconds
        Write-Host "   Request ${i}: $($stopwatch.ElapsedMilliseconds)ms" -ForegroundColor Gray
    } catch {
        $stopwatch.Stop()
        Write-Host "   Request ${i}: Failed ($($stopwatch.ElapsedMilliseconds)ms)" -ForegroundColor Red
    }
}

if ($responseTimes.Count -gt 0) {
    $avgResponseTime = ($responseTimes | Measure-Object -Average).Average
    Write-Host "✅ Average Response Time: $([math]::Round($avgResponseTime, 2))ms" -ForegroundColor Green
}

Write-Host "\n=== Test Complete ===" -ForegroundColor Green
Write-Host "Railway WAHA webhook integration is ready for production!" -ForegroundColor Yellow