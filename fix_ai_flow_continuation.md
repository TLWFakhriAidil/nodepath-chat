# AI Flow Continuation Fix

## Issue Identified
When an AI prompt node finishes processing and the next node is a user_reply node, the system correctly advances to the user_reply node and sets the waiting flag. However, when the user actually sends their reply, the system is not properly continuing the flow from the user_reply node to the subsequent nodes.

## Root Cause
The flow termination happens because:
1. After AI prompt completes, it advances to user_reply node and waits ✅
2. When user replies, it processes at the user_reply node but doesn't advance to the NEXT node after user_reply
3. The processUserReplyNode function only sets waiting state but doesn't actually process the user's input to move forward

## Solution
We need to modify the flow processing logic to:
1. When at a user_reply node with actual user input, process that input and advance to the next node
2. Ensure the flow continues after user_reply nodes instead of stopping

## Files to Modify
- `internal/whatsapp/whatsapp_service.go`
  - Fix `processUserReplyNode` function
  - Fix `processWaitingReplyTimesNode` function  
  - Update flow continuation logic

## Implementation Plan
1. Modify processUserReplyNode to check if we have user input
2. If we have user input, advance to the next node and continue processing
3. If no user input, set waiting state (current behavior)
4. Apply same logic to waiting_reply_times nodes
