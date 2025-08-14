import re

# Read the file
with open('src/ui/rest/app.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Fix pattern 1: Line 3382 area - missing closing brace
pattern1 = r'(} else if status == "failed" \{\s*query \+= ` HAVING bm\.status IN \(\'failed\', \'error\'\)`)\s*\n\s*\n\s*(// Add date filter if provided)'
replacement1 = r'\1\n\t}\n}\n\n\2'

content = re.sub(pattern1, replacement1, content, flags=re.MULTILINE)

# Write the file back
with open('src/ui/rest/app.go', 'w', encoding='utf-8') as f:
    f.write(content)

print("Fixed missing closing braces after 'failed' status checks")
