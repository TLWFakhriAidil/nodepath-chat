-- Remove gmail and phone columns from user_nodepath table
ALTER TABLE user_nodepath 
DROP COLUMN gmail,
DROP COLUMN phone;