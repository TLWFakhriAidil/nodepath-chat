# MySQL Reserved Keywords Fix

## Problem
The application is encountering SQL syntax errors because `trigger` is a reserved keyword in MySQL. It needs to be escaped with backticks when used as a column name.

## Errors Fixed
1. Error in lead queries: `Error 1064 (42000): You have an error in your SQL syntax ... near 'trigger, created_at, updated_at'`
2. Error in broadcast cleanup: `Error 1064 (42000): You have an error in your SQL syntax ... near ')'`

## Files to Fix

### 1. Lead Repository (`src/repository/lead_repository.go`)
All references to the `trigger` column need to be escaped with backticks.

### 2. Queued Message Cleaner (`src/usecase/queued_message_cleaner.go`)
The SQL query has an extra parenthesis causing syntax error.

## Solutions Applied
