# Enhanced AI Response System - Implementation Guide

## Overview
This document outlines the enhancements needed to align the NodePath Chat AI system with the advanced PHP implementation, including onemessage combining, stage detection, and improved response handling.

## Key Features to Implement

### 1. Time-Based Throttling
- Check time difference before processing (4-second minimum)
- Prevent rapid consecutive responses
- Store last response timestamp in `balas` field

### 2. Stage Detection from User Input
- Extract stage from user messages using regex: `/\bstage\s*:\s*(.+)/i`
- Auto-update conversation stage when detected
- Maintain stage continuity throughout conversation

### 3. Advanced Prompt Structure
- Include behavior and closing prompts
- Add instructions for onemessage handling
- Implement stage-aware responses
- Prevent repetitive responses

### 4. Onemessage Combining Logic
- Detect `Jenis: "onemessage"` in response items
- Combine consecutive onemessage items
- Send as single message with newlines
- Log as BOT_COMBINED in conversation history

### 5. Enhanced Response Formats
Support multiple response formats:
- Format 1: Standard JSON with Stage and Response array
- Format 2: Legacy format with Stage: and Response: prefixes
- Format 3: Plain text fallback
- Format 4: JSON wrapped in triple backticks
- Format 5: Encapsulated JSON within response content

## Implementation Files to Modify

### 1. `internal/services/ai_whatsapp_service.go`
- Add time throttling logic
- Implement stage detection from user input
- Enhanced prompt building
- Onemessage combining algorithm

### 2. `internal/models/models.go`
- Add `Balas` field for timestamp tracking
- Add `Jenis` field support in response items

### 3. `internal/handlers/device_settings_handlers.go`
- Update webhook processing for stage detection
- Add time-based throttling check

## Detailed Implementation Code