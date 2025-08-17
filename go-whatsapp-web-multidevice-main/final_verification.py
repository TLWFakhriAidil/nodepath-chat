import re

def final_verification_and_cleanup():
    """Final check and cleanup for GetPendingMessagesAndLock usage"""
    
    print("FINAL VERIFICATION: Ensuring GetPendingMessagesAndLock is used everywhere")
    print("="*80)
    
    # 1. Check optimized_broadcast_processor.go
    print("\n1. Checking optimized_broadcast_processor.go...")
    with open('src/usecase/optimized_broadcast_processor.go', 'r') as f:
        content = f.read()
    
    if 'GetPendingMessagesAndLock' in content and 'GetPendingMessages(' not in content.replace('GetPendingMessagesAndLock', ''):
        print("   ✓ CORRECT: Using GetPendingMessagesAndLock")
    else:
        print("   ✗ ERROR: Still using GetPendingMessages")
    
    # 2. Check broadcast_worker_processor.go
    print("\n2. Checking broadcast_worker_processor.go...")
    with open('src/usecase/broadcast_worker_processor.go', 'r') as f:
        content = f.read()
    
    if 'GetPendingMessagesAndLock' in content and 'GetPendingMessages(' not in content.replace('GetPendingMessagesAndLock', ''):
        print("   ✓ CORRECT: Using GetPendingMessagesAndLock")
    else:
        print("   ✗ ERROR: Still using GetPendingMessages")
    
    # 3. Remove old GetPendingMessages function
    print("\n3. Removing old GetPendingMessages function from broadcast_repository.go...")
    with open('src/repository/broadcast_repository.go', 'r') as f:
        lines = f.readlines()
    
    # Find and remove the old function
    new_lines = []
    skip = False
    removed = False
    
    for i, line in enumerate(lines):
        if '// GetPendingMessages gets pending messages' in line and 'GetPendingMessagesAndLock' not in line:
            skip = True
            removed = True
            continue
        
        if skip and line.strip() == '}' and i+1 < len(lines) and lines[i+1].strip() == '':
            skip = False
            continue
            
        if not skip:
            new_lines.append(line)
    
    if removed:
        with open('src/repository/broadcast_repository.go', 'w') as f:
            f.writelines(new_lines)
        print("   ✓ Removed old GetPendingMessages function")
    else:
        print("   - Old function not found or already removed")
    
    # 4. Verify GetPendingMessagesAndLock implementation
    print("\n4. Verifying GetPendingMessagesAndLock implementation...")
    with open('src/repository/broadcast_repository.go', 'r') as f:
        content = f.read()
    
    if 'processing_worker_id = ?' in content and 'processing_started_at = NOW()' in content:
        print("   ✓ Worker ID and timestamp are being set correctly")
    else:
        print("   ✗ ERROR: Worker ID implementation missing!")
    
    # 5. Check for any remaining calls to GetPendingMessages
    print("\n5. Checking for any remaining GetPendingMessages calls...")
    import os
    found_old_calls = False
    
    for root, dirs, files in os.walk('src'):
        # Skip backup files
        if 'backup' in root:
            continue
            
        for file in files:
            if file.endswith('.go') and 'backup' not in file:
                filepath = os.path.join(root, file)
                with open(filepath, 'r') as f:
                    content = f.read()
                
                # Look for old function calls
                if 'GetPendingMessages(' in content and 'GetPendingMessagesAndLock' not in content:
                    print(f"   ✗ Found old call in: {filepath}")
                    found_old_calls = True
    
    if not found_old_calls:
        print("   ✓ No old GetPendingMessages calls found")
    
    print("\n" + "="*80)
    print("VERIFICATION COMPLETE")
    print("="*80)

if __name__ == "__main__":
    final_verification_and_cleanup()
