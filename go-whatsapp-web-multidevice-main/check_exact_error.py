with open('src/ui/rest/app.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()

# Check line 3480, column 63
if len(lines) > 3479:
    line_3480 = lines[3479]
    print(f"Line 3480 (length {len(line_3480)}):")
    print(repr(line_3480))
    if len(line_3480) > 62:
        print(f"Character at column 63: {repr(line_3480[62])}")
        print(f"Context around column 63: {repr(line_3480[55:70])}")

print()

# Check line 3605, column 51  
if len(lines) > 3604:
    line_3605 = lines[3604]
    print(f"Line 3605 (length {len(line_3605)}):")
    print(repr(line_3605))
    if len(line_3605) > 50:
        print(f"Character at column 51: {repr(line_3605[50])}")
        print(f"Context around column 51: {repr(line_3605[45:60])}")
