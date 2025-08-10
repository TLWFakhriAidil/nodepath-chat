with open('src/ui/rest/app.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()

# Check around line 3480
print("Around line 3480:")
for i in range(3475, 3485):
    if i < len(lines):
        print(f"{i+1}: {lines[i].rstrip()}")

print("\nAround line 3605:")
for i in range(3600, 3610):
    if i < len(lines):
        print(f"{i+1}: {lines[i].rstrip()}")
