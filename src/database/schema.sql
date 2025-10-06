-- Create media_files table for storing file metadata
CREATE TABLE IF NOT EXISTS media_files (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  filename TEXT NOT NULL,
  original_name TEXT NOT NULL,
  file_type TEXT NOT NULL,
  file_size BIGINT NOT NULL,
  storage_path TEXT NOT NULL,
  public_url TEXT NOT NULL,
  uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create storage bucket for media files
INSERT INTO storage.buckets (id, name, public)
VALUES ('media', 'media', true)
ON CONFLICT (id) DO NOTHING;

-- Set up Row Level Security (RLS)
ALTER TABLE media_files ENABLE ROW LEVEL SECURITY;

-- Create policies for media_files table
-- Allow all operations for authenticated users (adjust as needed)
CREATE POLICY "Allow all operations for authenticated users" ON media_files
  FOR ALL USING (auth.role() = 'authenticated');

-- Create policy for public read access to storage
CREATE POLICY "Allow public read access" ON storage.objects
  FOR SELECT USING (bucket_id = 'media');

-- Create policy for authenticated upload to storage
CREATE POLICY "Allow authenticated upload" ON storage.objects
  FOR INSERT WITH CHECK (bucket_id = 'media' AND auth.role() = 'authenticated');

-- Create policy for authenticated delete from storage
CREATE POLICY "Allow authenticated delete" ON storage.objects
  FOR DELETE USING (bucket_id = 'media' AND auth.role() = 'authenticated');

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_media_files_uploaded_at ON media_files(uploaded_at DESC);
CREATE INDEX IF NOT EXISTS idx_media_files_file_type ON media_files(file_type);
CREATE INDEX IF NOT EXISTS idx_media_files_filename ON media_files(filename);

-- Create updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_media_files_updated_at 
    BEFORE UPDATE ON media_files 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- ========================================
-- BILLING AND SUBSCRIPTION TABLES
-- ========================================

-- Create subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  user_id UUID NOT NULL, -- Reference to user/customer
  plan_name TEXT NOT NULL DEFAULT 'Test Plan',
  plan_price DECIMAL(10,2) NOT NULL DEFAULT 1.00,
  plan_period TEXT NOT NULL DEFAULT 'monthly',
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled', 'suspended', 'pending')),
  next_billing_date DATE NOT NULL,
  features JSONB DEFAULT '["WhatsApp Bot Integration", "Flow Builder Access", "Basic Analytics", "Standard Support"]',
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create payments table (for Billplz payment receipts)
CREATE TABLE IF NOT EXISTS payments (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
  user_id UUID NOT NULL,
  bill_id TEXT, -- Billplz bill ID
  invoice_number TEXT UNIQUE NOT NULL, -- Our internal invoice number
  amount DECIMAL(10,2) NOT NULL,
  currency TEXT NOT NULL DEFAULT 'MYR',
  description TEXT,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'failed', 'cancelled')),
  payment_method TEXT DEFAULT 'billplz',
  billplz_url TEXT, -- Payment URL from Billplz
  paid_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create billing_history table (for display purposes)
CREATE TABLE IF NOT EXISTS billing_history (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  user_id UUID NOT NULL,
  payment_id UUID REFERENCES payments(id) ON DELETE CASCADE,
  invoice_number TEXT NOT NULL,
  amount DECIMAL(10,2) NOT NULL,
  currency TEXT NOT NULL DEFAULT 'MYR',
  description TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('pending', 'paid', 'failed', 'cancelled')),
  payment_date DATE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_next_billing ON subscriptions(next_billing_date);

CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_subscription_id ON payments(subscription_id);
CREATE INDEX IF NOT EXISTS idx_payments_bill_id ON payments(bill_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_billing_history_user_id ON billing_history(user_id);
CREATE INDEX IF NOT EXISTS idx_billing_history_payment_date ON billing_history(payment_date DESC);
CREATE INDEX IF NOT EXISTS idx_billing_history_status ON billing_history(status);

-- Create updated_at triggers
CREATE TRIGGER update_subscriptions_updated_at 
    BEFORE UPDATE ON subscriptions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payments_updated_at 
    BEFORE UPDATE ON payments 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Function to generate invoice numbers
CREATE OR REPLACE FUNCTION generate_invoice_number()
RETURNS TEXT AS $$
DECLARE
    invoice_num TEXT;
BEGIN
    invoice_num := 'INV-' || TO_CHAR(NOW(), 'YYYYMMDD') || '-' || LPAD(nextval('invoice_sequence')::TEXT, 4, '0');
    RETURN invoice_num;
END;
$$ LANGUAGE plpgsql;

-- Create sequence for invoice numbers
CREATE SEQUENCE IF NOT EXISTS invoice_sequence START 1;

-- Trigger to auto-generate invoice numbers
CREATE OR REPLACE FUNCTION set_invoice_number()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.invoice_number IS NULL OR NEW.invoice_number = '' THEN
        NEW.invoice_number := generate_invoice_number();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_payment_invoice_number
    BEFORE INSERT ON payments
    FOR EACH ROW
    EXECUTE FUNCTION set_invoice_number();

-- Function to update billing_history when payment status changes
CREATE OR REPLACE FUNCTION update_billing_history()
RETURNS TRIGGER AS $$
BEGIN
    -- Insert or update billing history record
    INSERT INTO billing_history (user_id, payment_id, invoice_number, amount, currency, description, status, payment_date)
    VALUES (
        NEW.user_id,
        NEW.id,
        NEW.invoice_number,
        NEW.amount,
        NEW.currency,
        NEW.description,
        NEW.status,
        CASE WHEN NEW.status = 'paid' THEN NEW.paid_at::DATE ELSE NULL END
    )
    ON CONFLICT (payment_id) DO UPDATE SET
        status = NEW.status,
        payment_date = CASE WHEN NEW.status = 'paid' THEN NEW.paid_at::DATE ELSE NULL END,
        amount = NEW.amount,
        description = NEW.description;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_billing_history_trigger
    AFTER INSERT OR UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION update_billing_history();

-- Insert default test subscription (for testing)
INSERT INTO subscriptions (
    user_id, 
    plan_name, 
    plan_price, 
    plan_period, 
    status, 
    next_billing_date,
    features
) VALUES (
    '00000000-0000-0000-0000-000000000001', -- Test user ID
    'Test Plan',
    1.00,
    'monthly',
    'active',
    (NOW() + INTERVAL '1 month')::DATE,
    '["WhatsApp Bot Integration", "Flow Builder Access", "Basic Analytics", "Standard Support"]'
) ON CONFLICT DO NOTHING;