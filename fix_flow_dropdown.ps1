$content = Get-Content "src\components\ChatbotBuilder.tsx"

# Find and replace the specific lines
for ($i = 0; $i -lt $content.Length; $i++) {
    # Replace flow name input with dropdown (line ~501)
    if ($content[$i] -match 'placeholder="Flow name\.\.\."') {
        $content[$i] = '                  <Select value={flowName} onValueChange={setFlowName}>'
        $content[$i+1] = '                    <SelectTrigger className="h-8 text-xs bg-background/50 border-border/50 focus:border-primary/50">'
        $content[$i+2] = '                      <SelectValue placeholder="Select flow type..." />'
        $content[$i+3] = '                    </SelectTrigger>'
        $content[$i+4] = '                    <SelectContent>'
        
        # Insert new lines for select items
        $newLines = @(
            '                      <SelectItem value="WasapBot Exama">WasapBot Exama</SelectItem>',
            '                      <SelectItem value="Chatbot AI">Chatbot AI</SelectItem>',
            '                    </SelectContent>',
            '                  </Select>'
        )
        
        # Replace the closing /> with our new lines
        $content[$i+5] = $newLines[0]
        
        # Insert remaining new lines
        $before = $content[0..($i+5)]
        $after = $content[($i+6)..($content.Length-1)]
        $content = $before + $newLines[1..3] + $after
        
        break
    }
}

# Fix validation check
for ($i = 0; $i -lt $content.Length; $i++) {
    if ($content[$i] -match 'if \(!flowName\.trim\(\)\)') {
        $content[$i] = '    if (!flowName || flowName === "") {'
    }
    if ($content[$i] -match 'title: "Flow name required"') {
        $content[$i] = '        title: "Flow type required",'
    }
    if ($content[$i] -match 'description: "Please enter a name for your flow"') {
        $content[$i] = '        description: "Please select a flow type (WasapBot Exama or Chatbot AI)",'
    }
}

$content | Out-File "src\components\ChatbotBuilder.tsx" -Encoding UTF8