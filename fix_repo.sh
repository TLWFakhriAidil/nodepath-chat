#!/bin/bash
# Fix script for removed database columns

# Fix ai_whatsapp_repository.go
sed -i 's/if ai\.FlowReference\.Valid {/\/\/ if ai.FlowReference.Valid {/g' internal/repository/ai_whatsapp_repository.go
sed -i 's/flowReferenceValue = ai\.FlowReference\.String/\/\/ flowReferenceValue = ai.FlowReference.String/g' internal/repository/ai_whatsapp_repository.go
sed -i 's/} else {/\/\/ } else {/g' internal/repository/ai_whatsapp_repository.go

sed -i 's/if ai\.ExecutionID\.Valid {/\/\/ if ai.ExecutionID.Valid {/g' internal/repository/ai_whatsapp_repository.go
sed -i 's/executionIDValue = ai\.ExecutionID\.String/\/\/ executionIDValue = ai.ExecutionID.String/g' internal/repository/ai_whatsapp_repository.go

sed -i 's/if ai\.ExecutionStatus\.Valid {/\/\/ if ai.ExecutionStatus.Valid {/g' internal/repository/ai_whatsapp_repository.go
sed -i 's/executionStatusValue = ai\.ExecutionStatus\.String/\/\/ executionStatusValue = ai.ExecutionStatus.String/g' internal/repository/ai_whatsapp_repository.go

sed -i 's/&ai\.ExecutionStatus, &ai\.ExecutionID,/\/\/ &ai.ExecutionStatus, &ai.ExecutionID,/g' internal/repository/ai_whatsapp_repository.go