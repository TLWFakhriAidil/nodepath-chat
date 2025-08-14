import re
import os

def final_verification_and_cleanup():
    """Final check and cleanup for GetPendingMessagesAndLock usage"""
    
    print("FINAL VERIFICATION: Ensuring GetPendingMessagesAndLock is used everywhere")
    print("="*80)
    
    issues_found = []
    
    # 1. Check optimized_broadcast_processor.go
    print("\n1. Checking optimized_broadcast_processor.go...")
    with open('src/usecase/optimized_broadcast_processor.go', 'r') as f:
        content = f.read()
    
    if 'GetPendingMessagesAndLock' in content and 'GetPendingMessages(' not in content.replace('GetPendingMessagesAndLock', ''):
        print("   [OK] Using GetPendingMessagesAndLock")
    else:
        print("   [ERROR] Still using GetPendingMessages")
        issues_found.append("optimized_broadcast_processor.go not using GetPendingMessagesAndLock")
    
    # 2. Check broadcast_worker_processor.go
    print("\n2. Checking broadcast_worker_processor.go...")
    with open('src/usecase/broadcast_worker_processor.go', 'r') as f:
        content = f.read()
    
    if 'GetPendingMessagesAndLock' in content and 'GetPendingMessages(' not in content.replace('GetPendingMessagesAndLock', ''):
        print("   [OK] Using GetPendingMessagesAndLock")
    else:
        print("   [ERROR] Still using GetPendingMessages")
        issues_found.append("broadcast_worker_processor.go not using GetPendingMessagesAndLock")
    
    # 3. Verify GetPendingMessagesAndLock implementation
    print("\n3. Verifying GetPendingMessagesAndLock implementation...")
    with open('src/repository/broadcast_repository.go', 'r') as f:
        content = f.read()
    
    if 'processing_worker_id = ?' in content and 'processing_started_at = NOW()' in content:
        print("   [OK] Worker ID and timestamp are being set correctly")
    else:
        print("   [ERROR] Worker ID implementation missing!")
        issues_found.append("GetPendingMessagesAndLock not setting worker ID properly")
    
    # 4. Check for any remaining calls to GetPendingMessages
    print("\n4. Checking for any remaining GetPendingMessages calls...")
    found_old_calls = []
    
    for root, dirs, files in os.walk('src'):
        # Skip backup files
        if 'backup' in root:
            continue
            
        for file in files:
            if file.endswith('.go') and 'backup' not in file:
                filepath = os.path.join(root, file)
                try:
                    with open(filepath, 'r') as f:
                        content = f.read()
                    
                    # Look for calls to GetPendingMessages (not GetPendingMessagesAndLock)
                    if re.search(r'\.GetPendingMessages\s*\(', content):
                        if 'GetPendingMessagesAndLock' not in content:
                            found_old_calls.append(filepath)
                            print(f"   [ERROR] Found old call in: {filepath}")
                except:
                    pass
    
    if not found_old_calls:
        print("   [OK] No old GetPendingMessages calls found in active files")
    else:
        issues_found.extend([f"Old GetPendingMessages call in {f}" for f in found_old_calls])
    
    # 5. Check if old GetPendingMessages function still exists
    print("\n5. Checking if old GetPendingMessages function exists...")
    with open('src/repository/broadcast_repository.go', 'r') as f:
        content = f.read()
    
    if re.search(r'func.*GetPendingMessages\s*\(.*\)\s*\(.*error.*\)', content) and 'GetPendingMessagesAndLock' not in re.search(r'func.*GetPendingMessages\s*\(.*\)\s*\(.*error.*\)', content).group():
        print("   [WARNING] Old GetPendingMessages function still exists")
        print("   Consider removing it to prevent confusion")
    else:
        print("   [OK] Old function not found or already removed")
    
    print("\n" + "="*80)
    print("VERIFICATION COMPLETE")
    print("="*80)
    
    if issues_found:
        print("\nISSUES FOUND:")
        for issue in issues_found:
            print(f"  - {issue}")
        return False
    else:
        print("\nALL CHECKS PASSED!")
        print("GetPendingMessagesAndLock is being used correctly everywhere.")
        return True

if __name__ == "__main__":
    all_good = final_verification_and_cleanup()
    
    if all_good:
        print("\nReady to build and push to GitHub!")
    else:
        print("\nPlease fix the issues before pushing.")
