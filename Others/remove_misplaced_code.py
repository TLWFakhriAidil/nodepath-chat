import re

# Read the file
with open('src/ui/rest/app.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()

# Find and remove the misplaced date filter code (lines 3386-3396 approximately)
# We need to identify it precisely
new_lines = []
skip_lines = False
skip_count = 0

for i, line in enumerate(lines):
    # Start skipping when we see the misplaced date filter outside function
    if i > 3380 and i < 3400 and '// Add date filter if provided' in line and lines[i-1].strip() == '}' and lines[i-2].strip() == '}':
        skip_lines = True
        skip_count = 0
        continue
    
    # Skip the date filter block (about 15 lines)
    if skip_lines:
        skip_count += 1
        if skip_count > 15 or 'query += ` ORDER BY' in line:
            skip_lines = False
            # Don't skip this line if it's the ORDER BY
            if 'query += ` ORDER BY' in line:
                new_lines.append(line)
        continue
    
    new_lines.append(line)

# Write back
with open('src/ui/rest/app.go', 'w', encoding='utf-8') as f:
    f.writelines(new_lines)

print("Removed misplaced date filter code")
print(f"Original lines: {len(lines)}, New lines: {len(new_lines)}")
