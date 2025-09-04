$content = Get-Content "src\components\ChatbotBuilder.tsx" -Raw

# Replace the Input field with Select dropdown
$content = $content -replace '(\s+)<Input\s+placeholder="Flow name\.\.\."[\s\S]*?onChange=\{[^}]+\}\)\}[\s\S]*?className="[^"]+"\s*/>', @'
$1<Select value={flowName} onValueChange={setFlowName}>
$1  <SelectTrigger className="h-8 text-xs bg-background/50 border-border/50 focus:border-primary/50">
$1    <SelectValue placeholder="Select flow type..." />
$1  </SelectTrigger>
$1  <SelectContent>
$1    <SelectItem value="WasapBot Exama">WasapBot Exama</SelectItem>
$1    <SelectItem value="Chatbot AI">Chatbot AI</SelectItem>
$1  </SelectContent>
$1</Select>
'@

# Update validation check
$content = $content -replace 'if \(!flowName\.trim\(\)\)', 'if (!flowName || flowName === "")'

# Update error messages
$content = $content -replace 'title: "Flow name required"', 'title: "Flow type required"'
$content = $content -replace 'description: "Please enter a name for your flow"', 'description: "Please select a flow type (WasapBot Exama or Chatbot AI)"'

$content | Set-Content "src\components\ChatbotBuilder.tsx" -NoNewline