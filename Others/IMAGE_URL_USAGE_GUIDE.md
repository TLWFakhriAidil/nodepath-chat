# Using External URLs for Images

## How to Use Images in Sequences and Campaigns

### 1. Upload to Your Server First
Upload images to growrvsb.com or any image hosting service:
- FTP to: `/home/admin/public_html/public/images/`
- Or use your Laravel upload feature
- Or use any image hosting (Imgur, Cloudinary, etc.)

### 2. Get the Full URL
Example URLs:
- `http://growrvsb.com/public/images/campaign/banner.jpg`
- `https://i.imgur.com/abc123.jpg`
- `https://res.cloudinary.com/your-cloud/image/upload/campaign.jpg`

### 3. Paste URL in WhatsApp System
- For Sequences: Paste in "Image URL" field
- For Campaigns: Paste in "Campaign Image URL" field
- For AI Campaigns: Paste in "Campaign Image URL" field

### Important Notes:
- URL must end with image extension (.jpg, .png, .gif, .webp)
- URL must be publicly accessible (no login required)
- URL is stored in database as-is (no base64 conversion)
- WhatsApp will fetch the image when sending

### Benefits:
- No storage on Railway server
- Images persist forever
- Faster loading (no base64 encoding)
- Can update images without changing campaigns
