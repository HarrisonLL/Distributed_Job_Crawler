-- Create a new role for user table access only
CREATE ROLE user_manager WITH LOGIN PASSWORD 'secure_password';

-- Grant connect permission to the database
GRANT CONNECT ON DATABASE gs_db TO user_manager;

-- Grant usage on schema (assuming public schema)
GRANT USAGE ON SCHEMA public TO user_manager;

-- Grant full access on user table only
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.users TO user_manager;
