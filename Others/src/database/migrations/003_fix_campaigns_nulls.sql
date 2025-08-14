-- Add ALTER statements to fix nullable columns
ALTER TABLE campaigns ALTER COLUMN niche SET DEFAULT '';
ALTER TABLE campaigns ALTER COLUMN image_url SET DEFAULT '';
ALTER TABLE campaigns ALTER COLUMN scheduled_time SET DEFAULT CURRENT_TIME;

-- Update existing NULL values
UPDATE campaigns SET niche = '' WHERE niche IS NULL;
UPDATE campaigns SET image_url = '' WHERE image_url IS NULL;
UPDATE campaigns SET scheduled_time = CURRENT_TIME WHERE scheduled_time IS NULL;
