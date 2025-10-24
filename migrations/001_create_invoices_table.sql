-- Migration: Create invoices table
-- This migration creates the invoices table with all required fields
-- Run this migration to add invoice functionality to the system

-- Create invoices table
CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number VARCHAR(255) NOT NULL UNIQUE,
    account_id VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL, -- Amount in cents to avoid floating point issues
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    due_date TIMESTAMPTZ NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, paid, failed, void
    metadata JSONB DEFAULT '{}',
    paid_at TIMESTAMPTZ NULL,
    payment_id VARCHAR(255) NULL, -- Reference to payment that paid this invoice
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add foreign key constraint to company_profiles table
ALTER TABLE invoices 
ADD CONSTRAINT fk_invoices_account_id 
FOREIGN KEY (account_id) REFERENCES company_profiles(id) ON DELETE CASCADE;

-- Add unique index to prevent duplicate invoices per account per due date
CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_account_due_date 
ON invoices (account_id, due_date) 
WHERE status != 'void';

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_invoices_account_id ON invoices (account_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices (status);
CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices (due_date);
CREATE INDEX IF NOT EXISTS idx_invoices_issued_at ON invoices (issued_at);

-- Add trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_invoices_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_invoices_updated_at
    BEFORE UPDATE ON invoices
    FOR EACH ROW
    EXECUTE FUNCTION update_invoices_updated_at();

-- Add comments for documentation
COMMENT ON TABLE invoices IS 'Stores invoice records for subscription billing';
COMMENT ON COLUMN invoices.id IS 'Primary key UUID';
COMMENT ON COLUMN invoices.invoice_number IS 'Unique invoice number (e.g., INV-2024-001)';
COMMENT ON COLUMN invoices.account_id IS 'Foreign key to company_profiles.id';
COMMENT ON COLUMN invoices.amount IS 'Invoice amount in cents (e.g., 5000 = $50.00)';
COMMENT ON COLUMN invoices.currency IS 'Currency code (USD, EUR, etc.)';
COMMENT ON COLUMN invoices.due_date IS 'Date when payment is due';
COMMENT ON COLUMN invoices.issued_at IS 'Date when invoice was created';
COMMENT ON COLUMN invoices.status IS 'Invoice status: pending, paid, failed, void';
COMMENT ON COLUMN invoices.metadata IS 'Additional invoice data (plan details, etc.)';
COMMENT ON COLUMN invoices.paid_at IS 'Date when invoice was paid';
COMMENT ON COLUMN invoices.payment_id IS 'Reference to payment record that paid this invoice';
