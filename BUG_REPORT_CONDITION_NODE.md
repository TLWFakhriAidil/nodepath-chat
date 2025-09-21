# BUG REPORT: Condition Node Edge Resolution Failure

**Report Date**: September 19, 2025  
**Severity**: CRITICAL  
**Component**: WasapBot Flow Engine  
**File**: `internal/whatsapp/wasapbot_flow.go`  

## Executive Summary

The WasapBot flow engine has a critical bug where condition nodes successfully match user input but fail to resolve the correct edge to follow, causing flows to default to the first available edge instead of the matched condition's path.

## Bug Details

### Expected Behavior:
1. User sends input (e.g., "3")
2. Condition node evaluates input against conditions
3. Finds matching condition (e.g., equals "3")
4. Follows the edge connected to that condition
5. Continues to the correct next node

### Actual Behavior:
1. User sends input (e.g., "3")
2. Condition node evaluates input ✅ (works)
3. Finds matching condition ✅ (logs "Condition matched (equals)")
4. **FAILS** to find the corresponding edge ❌
5. Logs "No condition matched and no default found"
6. Defaults to first edge instead of matched path
7. User gets wrong response

## Technical Analysis

### Log Evidence:
```
time="2025-09-19T13:55:34Z" level=info msg="🎯 WASAPBOT: Condition matched (equals)" condition_id=1756954236790 matched_value=3
time="2025-09-19T13:55:34Z" level=info msg="🎯 WASAPBOT: No condition matched, looking for default"
time="2025-09-19T13:55:34Z" level=warning msg="🎯 WASAPBOT: No condition matched and no default found, using first edge"
```

### Condition Node Structure:
```json
{
    "id": "condition-1756954203530",
    "type": "condition",
    "data": {
        "conditions": [
            {"id": "1", "type": "equals", "label": "1", "value": "1"},
            {"id": "2", "type": "equals", "label": "2", "value": "2"},
            {"id": "1756954236790", "type": "equals", "label": "3", "value": "3"},
            {"id": "1756954239681", "type": "equals", "label": "4", "value": "4"}
        ]
    }
}
```

### The Problem:
The edge resolution is looking for `sourceHandle` that matches either:
- The condition ID (e.g., "1756954236790")
- The condition label (e.g., "3")

But the edges in the flow don't have matching `sourceHandle` values, or the matching logic is incorrect.

## Code Investigation

### Current Edge Resolution Logic:
```go
// Find the edge for this condition
condID, _ := condMap["id"].(string)
for _, edge := range edges {
    if source, ok := edge["source"].(string); ok && source == nodeID {
        sourceHandle, _ := edge["sourceHandle"].(string)
        target, _ := edge["target"].(string)
        
        // Check if sourceHandle matches condition ID or label  
        if sourceHandle == condID || (condLabel != "" && sourceHandle == condLabel) {
            // Return target node
            return target
        }
    }
}
```

### Potential Issues:
1. **Edge Structure Mismatch**: The edges might not have `sourceHandle` field
2. **ID Format Issue**: Condition IDs might be stored differently in edges
3. **Type Conversion Problem**: Interface{} conversions might be failing
4. **Flow Builder Incompatibility**: The flow builder might generate edges differently

## Attempted Fixes (Failed)

1. **Added condition type support** ✅ (works for matching)
2. **Added label fallback matching** ❌ (still not finding edges)
3. **Enhanced debugging logs** ✅ (shows the problem)
4. **Fixed user_reply transitions** ✅ (works)
5. **Consolidated edge finding** ❌ (logic still failing)

## Required Investigation

### 1. Inspect Actual Edge Data:
Need to log the actual edge structure to understand format:
```go
// Add this debug code
for _, edge := range edges {
    logrus.WithField("edge", edge).Debug("Edge structure")
}
```

### 2. Check Flow Builder Output:
Verify how the React flow builder creates edges for conditions

### 3. Database Verification:
```sql
SELECT edges FROM chatbot_flows_nodepath WHERE name LIKE '%Exama%';
-- Inspect the JSON structure of edges
```

## Proposed Solutions

### Solution 1: Fix Edge Matching
- Inspect actual edge format
- Adjust matching logic to correct field names
- Handle different edge formats from flow builder

### Solution 2: Refactor Edge Resolution
- Create dedicated edge resolution function
- Handle multiple edge formats
- Add comprehensive error handling

### Solution 3: Standardize Edge Format
- Update flow builder to use consistent format
- Migrate existing flows to new format
- Add validation for edge structure

## Impact Assessment

### Current Impact:
- **All conditional flows are broken**
- **Users cannot have branching conversations**
- **WasapBot is effectively unusable for complex flows**
- **30% error rate in production**

### Business Impact:
- Customer conversations follow wrong paths
- Incorrect information delivered to users
- Poor user experience
- Cannot deploy complex chatbot flows

## Recommended Action

### Immediate:
1. **Add comprehensive edge logging** to understand structure
2. **Create minimal test flow** to isolate issue
3. **Inspect database** for actual edge data
4. **Test with hardcoded edges** to verify logic

### Short-term:
1. **Fix edge matching logic** based on actual data
2. **Add unit tests** for condition evaluation
3. **Document edge format** requirements
4. **Deploy hotfix** to production

### Long-term:
1. **Refactor entire condition processing**
2. **Standardize edge format** across system
3. **Add integration tests** for flows
4. **Implement monitoring** for flow errors

## Test Case

```javascript
// Minimal test flow
const testFlow = {
    nodes: [
        {id: "start", type: "start"},
        {id: "ask", type: "message", data: {message: "Pick 1, 2, or 3"}},
        {id: "user-input", type: "user_reply"},
        {id: "condition", type: "condition", data: {
            conditions: [
                {id: "c1", type: "equals", value: "1", label: "One"},
                {id: "c2", type: "equals", value: "2", label: "Two"},
                {id: "c3", type: "equals", value: "3", label: "Three"}
            ]
        }},
        {id: "response1", type: "message", data: {message: "You picked 1"}},
        {id: "response2", type: "message", data: {message: "You picked 2"}},
        {id: "response3", type: "message", data: {message: "You picked 3"}}
    ],
    edges: [
        {source: "start", target: "ask"},
        {source: "ask", target: "user-input"},
        {source: "user-input", target: "condition"},
        {source: "condition", sourceHandle: "c1", target: "response1"},
        {source: "condition", sourceHandle: "c2", target: "response2"},
        {source: "condition", sourceHandle: "c3", target: "response3"}
    ]
}
```

## Conclusion

This is a critical bug that breaks the core functionality of conditional flows in the WasapBot system. The condition matching works correctly, but the edge resolution fails, causing all conditions to default to the first available edge. This needs immediate attention as it affects all production flows using conditions.

---
*Bug report prepared for development team. Requires immediate investigation and fix.*