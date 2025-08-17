import re

def fix_queue_message():
    """Fix the QueueMessage function to use transactions for atomic duplicate checking"""
    
    # Read the file
    with open('src/repository/broadcast_repository.go', 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Find the QueueMessage function
    pattern = r'(func \(r \*BroadcastRepository\) QueueMessage\(msg domainBroadcast\.BroadcastMessage\) error \{[^}]+?// ISSUE 3 FIX: Check for duplicates before inserting)'
    match = re.search(pattern, content, re.DOTALL)
    
    if not match:
        print("Could not find QueueMessage function")
        return False
    
    # New function with transaction support
    new_function = '''func (r *BroadcastRepository) QueueMessage(msg domainBroadcast.BroadcastMessage) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	
	// Use transaction for atomic duplicate check and insert
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// ISSUE 3 FIX: Check for duplicates before inserting'''
    
    # Replace the function start
    content = content.replace(match.group(1), new_function)
    
    # Replace all r.db with tx in the QueueMessage function
    # Find the end of the function
    func_start = content.find('func (r *BroadcastRepository) QueueMessage')
    func_end = content.find('\n}\n\n// GetPendingMessages', func_start)
    
    if func_start == -1 or func_end == -1:
        print("Could not find function boundaries")
        return False
    
    # Extract the function
    func_content = content[func_start:func_end+2]
    
    # Replace r.db with tx in this function
    func_content = func_content.replace('r.db.QueryRow(duplicateCheck,', 'tx.QueryRow(duplicateCheck,')
    func_content = func_content.replace('r.db.QueryRow("SELECT user_id', 'tx.QueryRow("SELECT user_id')
    func_content = func_content.replace('r.db.Exec(query,', 'tx.Exec(query,')
    
    # Add tx.Commit() before the final return
    func_content = func_content.replace('\treturn err\n}', '\tif err != nil {\n\t\treturn err\n\t}\n\t\n\t// Commit the transaction\n\treturn tx.Commit()\n}')
    
    # Also add 'processing' status to duplicate checks
    func_content = func_content.replace("AND status IN ('pending', 'sent', 'queued')", "AND status IN ('pending', 'sent', 'queued', 'processing')")
    
    # Replace in the original content
    content = content[:func_start] + func_content + content[func_end+2:]
    
    # Write back
    with open('src/repository/broadcast_repository.go', 'w', encoding='utf-8') as f:
        f.write(content)
    
    print("Successfully updated QueueMessage to use transactions")
    return True

if __name__ == "__main__":
    fix_queue_message()
