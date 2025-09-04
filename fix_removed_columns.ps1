$basePath = "C:\Users\User\Documents\Trae\nodepath-chat-1"

# Fix ai_whatsapp_repository.go
$repoFile = "$basePath\internal\repository\ai_whatsapp_repository.go"
$content = Get-Content $repoFile -Raw

# Remove executionStatus and executionID from method signatures
$content = $content -replace 'UpdateFlowTrackingFields\(prospectNum, idDevice string, flowID, currentNodeID, lastNodeID string, waitingForReply int, executionStatus, executionID string\)', 'UpdateFlowTrackingFields(prospectNum, idDevice string, flowID, currentNodeID, lastNodeID string, waitingForReply int)'

# Remove FlowReference, ExecutionStatus, ExecutionID field handling
$content = $content -replace '\s+if ai\.FlowReference\.Valid \{[^}]+\} else \{[^}]+\}', ''
$content = $content -replace '\s+if ai\.ExecutionID\.Valid \{[^}]+\} else \{[^}]+\}', ''
$content = $content -replace '\s+if ai\.ExecutionStatus\.Valid \{[^}]+\} else \{[^}]+\}', ''

# Remove from Scan calls
$content = $content -replace ',\s*&ai\.ExecutionStatus,\s*&ai\.ExecutionID', ''

# Remove variable declarations and assignments for removed fields
$content = $content -replace 'var flowReferenceValue, executionIDValue, executionStatusValue interface\{\}', ''
$content = $content -replace '\s+flowReferenceValue = nil', ''
$content = $content -replace '\s+executionIDValue = nil', ''
$content = $content -replace '\s+executionStatusValue = nil', ''

[System.IO.File]::WriteAllText($repoFile, $content)
Write-Host "Fixed ai_whatsapp_repository.go"