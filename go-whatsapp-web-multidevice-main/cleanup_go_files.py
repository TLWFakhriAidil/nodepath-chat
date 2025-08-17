import os
import shutil

# List of Go files to move
go_files = [
    'fix_campaign_summary_direct.go',
    'fix_sequence_leads_calculation.go',
    'test_actual_flow.go',
    'test_correct_flow.go',
    'test_greeting_order.go',
    'test_malaysian_greeting.go',
    'test_name_timezone.go',
    'test_new_greeting_flow.go',
    'test_simple_name_clean.go',
    'sequence_date_filter_changes.go'
]

# Create old_files directory if it doesn't exist
if not os.path.exists('old_files'):
    os.makedirs('old_files')

# Move each file
for file in go_files:
    if os.path.exists(file):
        try:
            shutil.move(file, os.path.join('old_files', file))
            print(f"Moved {file} to old_files/")
        except Exception as e:
            print(f"Error moving {file}: {e}")
    else:
        print(f"File not found: {file}")

print("\nCleanup complete!")
