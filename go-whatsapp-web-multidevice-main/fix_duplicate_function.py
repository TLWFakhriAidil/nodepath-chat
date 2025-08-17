# Fix duplicate function by finding and removing the second occurrence

# Read the file
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository\broadcast_repository.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Find all occurrences of the function
import re

# Split by lines to find the duplicate
lines = content.split('\n')

# Find line numbers where function starts
func_starts = []
for i, line in enumerate(lines):
    if 'func (r *BroadcastRepository) GetPendingMessagesAndLock' in line:
        func_starts.append(i)

print(f"Found function declarations at lines: {func_starts}")

if len(func_starts) >= 2:
    # Find the end of first function (look for closing brace at start of line)
    first_func_end = None
    brace_count = 0
    in_function = False
    
    for i in range(func_starts[0], func_starts[1]):
        line = lines[i]
        if 'func (r *BroadcastRepository) GetPendingMessagesAndLock' in line:
            in_function = True
            brace_count = 0
        
        if in_function:
            brace_count += line.count('{') - line.count('}')
            if brace_count == 0 and '{' in line:
                # Function ended
                first_func_end = i + 1
                break
    
    if first_func_end:
        # Remove everything from end of first function to start of second
        del lines[first_func_end:func_starts[1]]
        
        # Write back
        with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository\broadcast_repository.go', 'w', encoding='utf-8') as f:
            f.write('\n'.join(lines))
        
        print(f"Removed duplicate function from line {first_func_end} to {func_starts[1]}")
    else:
        print("Could not find end of first function")
else:
    print("No duplicate found")
