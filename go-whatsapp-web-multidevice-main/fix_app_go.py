import re

# Read the file
with open('src/ui/rest/app.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Fix the malformed log.Printf line
content = content.replace(
    'GetSequenceStepLeads - Sequence: %s\n\tsequenceId, deviceId, stepId, status, startDate, endDate)',
    'log.Printf("GetSequenceStepLeads - Sequence: %s, Device: %s, Step: %s, Status: %s, DateRange: %s to %s",\n\t\tsequenceId, deviceId, stepId, status, startDate, endDate)'
)

# Fix the missing brace after the failed condition
pattern = r'} else if status == "failed" \{\s*query \+= ` AND bm\.status IN \(\'failed\', \'error\'\)`\s*\n\s*// Add date filter if provided'
replacement = '''} else if status == "failed" {
			query += ` AND bm.status IN ('failed', 'error')`
		}
	}
	
	// Add date filter if provided'''

content = re.sub(pattern, replacement, content, flags=re.MULTILINE)

# Write the file back
with open('src/ui/rest/app.go', 'w', encoding='utf-8') as f:
    f.write(content)

print("Fixed app.go successfully!")
