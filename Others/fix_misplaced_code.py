import re

# Read the file
with open('src/ui/rest/app.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Find and remove the misplaced date filter block
# This is around line 3386-3396
pattern = r'\n// Add date filter if provided\n\tif startDate != "" && endDate != "" \{[^}]+\}\n\tquery \+= ` ORDER BY bm\.sent_at DESC`'

# Search for the pattern
if re.search(pattern, content, re.DOTALL):
    print("Found misplaced code block")
    
# Remove the specific misplaced block
lines = content.split('\n')
new_lines = []
skip = False
skip_count = 0

for i, line in enumerate(lines):
    if i >= 3385 and i <= 3397 and '// Add date filter if provided' in line:
        skip = True
        skip_count = 0
        continue
    
    if skip:
        skip_count += 1
        if skip_count >= 12:  # Skip the misplaced block
            skip = False
        continue
        
    new_lines.append(line)

# Write back
with open('src/ui/rest/app.go', 'w', encoding='utf-8') as f:
    f.write('\n'.join(new_lines))

print("Fixed the misplaced code block")
