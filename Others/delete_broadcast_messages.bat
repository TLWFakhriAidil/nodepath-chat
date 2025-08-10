#!/bin/bash
# Delete all broadcast messages using psql command

psql "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway" -c "DELETE FROM broadcast_messages; SELECT COUNT(*) as remaining FROM broadcast_messages;"
