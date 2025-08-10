#!/usr/bin/env python3
import os
import re

# Find specific files mentioned in the search results
files_to_fix = [
    'src/usecase/sequence.go',
    'src/usecase/optimized_campaign_trigger.go',
    'src/usecase/campaign_trigger.go',
    'src/usecase/ai_campaign_processor.go',
    'src/ui/rest/app.go',
    'src/ui/rest/device_refresh.go',
    'src/repository/user_repository.go',
    'src/infrastructure/broadcast/stable_message_sender.go',
    'src/infrastructure/whatsapp/device_status_normalizer.go',
    'src/infrastructure/whatsapp/device_health_monitor.go',
    'src/infrastructure/whatsapp/multidevice/manager.go',
    'src/infrastructure/whatsapp/auto_connection_monitor_15min.go',
    'src/infrastructure/broadcast/whatsapp_message_sender.go'
]

for file_path in files_to_fix:
    if not os.path.exists(file_path):
        print(f"File not found: {file_path}")
        continue
    
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()
        
        modified = False
        new_lines = []
        
        for i, line in enumerate(lines):
            # Look for the broken pattern
            if line.strip() == 'Platform != ""':
                # Look at previous lines to understand context
                indent = len(line) - len(line.lstrip())
                new_lines.append(' ' * indent + 'if device.Platform != "" {\n')
                modified = True
                print(f"Fixed line {i+1} in {file_path}")
            elif 'Platform != ""' in line and 'if' not in line and line.strip().startswith('Platform'):
                # Another variation
                indent = len(line) - len(line.lstrip())
                new_lines.append(' ' * indent + 'if device.Platform != "" {\n')
                modified = True
                print(f"Fixed line {i+1} in {file_path}")
            else:
                new_lines.append(line)
        
        if modified:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.writelines(new_lines)
            print(f"Updated: {file_path}")
    
    except Exception as e:
        print(f"Error processing {file_path}: {e}")

print("\nDone!")