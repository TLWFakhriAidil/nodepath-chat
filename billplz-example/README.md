# Billplz Payment Integration

A secure PHP integration for Billplz payment gateway.

## ⚠️ IMPORTANT SECURITY NOTICE

**NEVER commit sensitive credentials to Git!**

## Setup Instructions

1. **Copy the configuration template:**
   ```bash
   cp config.example.php config.php
   ```

2. **Edit `config.php` with your actual credentials:**
   - Add your Billplz API key
   - Add your database credentials
   - Update URLs to match your domain

3. **Ensure `config.php` is in `.gitignore`** (already included)

4. **Database Setup:**
   ```sql
   CREATE TABLE orders (
       id INT AUTO_INCREMENT PRIMARY KEY,
       customer_email VARCHAR(255),
       customer_name VARCHAR(255),
       billing_phone VARCHAR(20),
       billing_address TEXT,
       billing_city VARCHAR(100),
       billing_state VARCHAR(100),
       billing_postcode VARCHAR(10),
       amount DECIMAL(10, 2),
       collection_id VARCHAR(50),
       status VARCHAR(50),
       bill_id VARCHAR(50),
       product VARCHAR(255),
       method VARCHAR(50),
       url TEXT,
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );

   CREATE TABLE abandoned_leads (
       id INT AUTO_INCREMENT PRIMARY KEY,
       billing_phone VARCHAR(20),
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );
   ```

## Files

- `billplz_payment.php` - Main payment processing
- `callback.php` - Webhook handler for Billplz
- `thank_you.php` - Success page after payment
- `config.example.php` - Configuration template (rename to config.php)

## Security Best Practices

1. **Use environment variables** for production
2. **Enable HTTPS** for all payment pages
3. **Validate all input** from forms and callbacks
4. **Log errors** but never log sensitive data
5. **Use prepared statements** for all database queries
6. **Implement rate limiting** on payment endpoints
7. **Verify Billplz signatures** on callbacks

## Testing

Use Billplz Sandbox for testing:
- Set `sandbox => true` in config.php
- Use sandbox API credentials
- Test with sandbox URLs

## Support

For Billplz API documentation: https://www.billplz.com/api