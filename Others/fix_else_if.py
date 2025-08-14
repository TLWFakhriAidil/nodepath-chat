import re

# Read the file
with open('src/ui/rest/app.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Fix the missing "} else" before "if status == "failed""
# Pattern to find lines where "} else" is missing
pattern = r'(\} else if status == "pending" \{[^\}]*query \+= ` (?:AND|HAVING) bm\.status[^\}]*\)\"`\s*\n\s*)(if status == "failed" \{)'

# Replace with proper else if
replacement = r'\1} else \2'

content = re.sub(pattern, replacement, content, flags=re.MULTILINE)

# Write the file back
with open('src/ui/rest/app.go', 'w', encoding='utf-8') as f:
    f.write(content)

print("Fixed missing '} else' before 'if status == \"failed\"'")

# Count occurrences to verify
occurrences = len(re.findall(r'} else if status == "failed"', content))
print(f"Found {occurrences} proper 'else if' statements for failed status")
