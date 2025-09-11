# Test script to verify prospect_name saving functionality
# Test parameters: id_device FakhriAidilTLW-001, flow flow_ai_1756016272, phone 601137508067

$webhook_url = "http://localhost:8080/api/ai-whatsapp/webhook/whatsapp/FakhriAidilTLW-001"

# Test webhook payload with prospect name
$payload = @{
    id_device = "FakhriAidilTLW-001"
    phone_number = "601137508067"
    sender_name = "Test User Fakhri"
    message = "Hello, this is a test message to verify prospect_name saving"
    flow = "flow_ai_1756016272"
    timestamp = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
} | ConvertTo-Json

Write-Host "Sending test webhook to verify prospect_name saving..."
Write-Host "Payload: $payload"

try {
    $response = Invoke-RestMethod -Uri $webhook_url -Method POST -Body $payload -ContentType "application/json"
    Write-Host "✅ Webhook sent successfully!"
    Write-Host "Response: $($response | ConvertTo-Json -Depth 3)"
} catch {
    Write-Host "❌ Error sending webhook: $($_.Exception.Message)"
    Write-Host "Response: $($_.Exception.Response | ConvertTo-Json -Depth 3)"
}

Write-Host "`nTest completed. Check the database to verify prospect_name was saved correctly."