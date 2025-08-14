import re

# Read the dashboard.html file
with open('src/views/dashboard.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Find the showSequenceStepLeadDetails function and update it
old_fetch_line = r"fetch\(`/api/sequences/\$\{currentSequenceForReport\.id\}/device/\$\{deviceId\}/step/\$\{stepId\}/leads\?status=\$\{status\}`, \{"

new_fetch_block = """// Build URL with date filters
            let url = `/api/sequences/${currentSequenceForReport.id}/device/${deviceId}/step/${stepId}/leads?status=${status}`;
            
            // Get the current date filters from the sequence summary
            const startDate = document.getElementById('sequenceStartDate').value;
            const endDate = document.getElementById('sequenceEndDate').value;
            
            if (startDate) {
                url += `&start_date=${startDate}`;
            }
            if (endDate) {
                url += `&end_date=${endDate}`;
            }
            
            console.log('Fetching sequence step leads with URL:', url);
            
            // Fetch lead details for specific step
            fetch(url, {"""

content = re.sub(
    r"// Fetch lead details for specific step\s*\n\s*fetch\(`/api/sequences/\$\{currentSequenceForReport\.id\}/device/\$\{deviceId\}/step/\$\{stepId\}/leads\?status=\$\{status\}`, \{",
    new_fetch_block,
    content
)

# Write the updated content back
with open('src/views/dashboard.html', 'w', encoding='utf-8') as f:
    f.write(content)

print("Updated showSequenceStepLeadDetails function in dashboard.html")
