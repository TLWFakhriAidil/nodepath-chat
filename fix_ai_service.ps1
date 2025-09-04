$basePath = "C:\Users\User\Documents\Trae\nodepath-chat-1"

# Fix ai_whatsapp_service.go
$serviceFile = "$basePath\internal\services\ai_whatsapp_service.go"
$content = Get-Content $serviceFile -Raw

# Remove StartFlowExecution method that uses removed fields
$content = $content -replace 'func \(s \*aiWhatsappService\) StartFlowExecution\(prospectNum, idDevice, flowReference string, variables map\[string\]interface\{\}\) \(\*models\.AIWhatsapp, error\) \{[\s\S]*?return aiConv, nil\s+\}', '// StartFlowExecution removed - uses deleted database columns'

# Remove UpdateFlowExecution method
$content = $content -replace 'func \(s \*aiWhatsappService\) UpdateFlowExecution\(prospectNum, idDevice, currentNode string, variables map\[string\]interface\{\}, status string\) error \{[\s\S]*?return nil\s+\}', '// UpdateFlowExecution removed - uses deleted database columns'

# Remove GetFlowExecutionVariables method
$content = $content -replace 'func \(s \*aiWhatsappService\) GetFlowExecutionVariables\(prospectNum, idDevice string\) \(map\[string\]interface\{\}, error\) \{[\s\S]*?return variables, nil\s+\}', '// GetFlowExecutionVariables removed - uses deleted database columns'

# Remove from interface
$content = $content -replace 'StartFlowExecution\(prospectNum, idDevice, flowReference string, variables map\[string\]interface\{\}\) \(\*models\.AIWhatsapp, error\)', '// StartFlowExecution removed'
$content = $content -replace 'UpdateFlowExecution\(prospectNum, idDevice, currentNode string, variables map\[string\]interface\{\}, status string\) error', '// UpdateFlowExecution removed'
$content = $content -replace 'GetFlowExecutionVariables\(prospectNum, idDevice string\) \(map\[string\]interface\{\}, error\)', '// GetFlowExecutionVariables removed'

# Update UpdateFlowTrackingFields calls to remove executionStatus and executionID parameters
$content = $content -replace 'r\.UpdateFlowTrackingFields\([^,]+,[^,]+,[^,]+,[^,]+,[^,]+,[^,]+,"[^"]*","[^"]*"\)', 'r.UpdateFlowTrackingFields($1,$2,$3,$4,$5,$6)'

[System.IO.File]::WriteAllText($serviceFile, $content)
Write-Host "Fixed ai_whatsapp_service.go"