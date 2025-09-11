-- Rollback migration - DEPRECATED COLUMNS REMOVED
-- Note: The deprecated columns (jam, catatan_staff, data_image, conv_stage, variables, bot_balas, current_node) 
-- are permanently removed and will not be restored in rollback to maintain schema consistency.

-- This rollback migration intentionally does nothing to prevent reinstalling deprecated columns
-- that have been removed from the new ai_whatsapp_nodepath schema.

SELECT 'Rollback migration completed - deprecated columns remain removed' AS status;
