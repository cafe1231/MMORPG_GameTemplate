-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    account_status INTEGER DEFAULT 1, -- 1=active, 2=suspended, 3=banned, 4=pending_verification, 5=deleted
    roles TEXT[] DEFAULT ARRAY['player'],
    max_characters INTEGER DEFAULT 5,
    character_count INTEGER DEFAULT 0,
    is_premium BOOLEAN DEFAULT FALSE,
    premium_expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT check_character_count CHECK (character_count >= 0 AND character_count <= max_characters),
    CONSTRAINT check_max_characters CHECK (max_characters > 0 AND max_characters <= 100)
);

-- Create indexes for performance
CREATE INDEX idx_users_email ON users(LOWER(email));
CREATE INDEX idx_users_username ON users(LOWER(username));
CREATE INDEX idx_users_account_status ON users(account_status);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE
    ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();