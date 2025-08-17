import re

# Read the file
with open('src/views/public_device.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Store original for comparison
original_content = content

# Fix 1: Function declarations without parentheses
# Pattern: function name\n or function name {
content = re.sub(r'function\s+(\w+)\s*\n', r'function \1()\n', content)
content = re.sub(r'function\s+(\w+)\s*{', r'function \1() {', content)

# Fix 2: Function declarations with duplicate parameters
# Pattern: function name() {() { or function name(params) {(params) {
content = re.sub(r'function\s+(\w+)\s*\(\)\s*{\s*\(\)\s*{', r'function \1() {', content)
content = re.sub(r'function\s+(\w+)\s*\([^)]*\)\s*{\s*\([^)]*\)\s*{', r'function \1() {', content)

# Fix 3: Extra () { patterns
content = re.sub(r'\)\s*{\s*\(\)\s*{', ') {', content)

# Fix 4: Look for specific known issues
# renderDevices() {() {
content = content.replace('renderDevices() {() {', 'renderDevices() {')

# loadSequenceSummary
content = content.replace('window.loadSequenceSummary = function(showAll = false) {() {', 
                         'window.loadSequenceSummary = function(showAll = false) {')

# displayCampaignSummary and displaySequenceSummary
content = content.replace('function displayCampaignSummary(summary) {(summary) {', 
                         'function displayCampaignSummary(summary) {')
content = content.replace('function displaySequenceSummary(summary) {(summary) {', 
                         'function displaySequenceSummary(summary) {')

# setDefaultDates
content = content.replace('function setDefaultDates() {() {', 'function setDefaultDates() {')

# Count braces to check balance
open_braces = content.count('{')
close_braces = content.count('}')
print(f"Brace count - Open: {open_braces}, Close: {close_braces}, Difference: {open_braces - close_braces}")

# Find all function declarations to verify they're correct
function_pattern = re.compile(r'function\s+(\w+)[^{]*{')
functions = function_pattern.findall(content)
print(f"\nFound {len(functions)} functions")

# Check for any remaining syntax issues
issues = []
lines = content.split('\n')
for i, line in enumerate(lines):
    # Check for function without parentheses
    if re.search(r'function\s+\w+\s*{', line) and '(' not in line:
        issues.append(f"Line {i+1}: Function missing parentheses: {line.strip()}")
    
    # Check for duplicate parameters
    if re.search(r'\)\s*{\s*\(.*\)\s*{', line):
        issues.append(f"Line {i+1}: Duplicate parameters: {line.strip()}")
    
    # Check for {() {
    if '{() {' in line:
        issues.append(f"Line {i+1}: Extra () after brace: {line.strip()}")

if issues:
    print("\nRemaining issues found:")
    for issue in issues:
        print(f"  {issue}")
else:
    print("\nNo syntax issues found!")

# Write the fixed content
if content != original_content:
    with open('src/views/public_device.html', 'w', encoding='utf-8') as f:
        f.write(content)
    print(f"\nFixed {len(original_content) - len(content)} characters")
    print("File has been updated!")
else:
    print("\nNo changes needed!")
