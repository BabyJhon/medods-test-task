-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS refresh_sessions (
    id UUID PRIMARY KEY NOT NULL,          
    user_id UUID NOT NULL,                 
    refresh_hash TEXT NOT NULL,            
    user_agent TEXT NOT NULL,             
    ip INET NOT NULL,                     
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),  
    expires_at TIMESTAMPTZ NOT NULL,      
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE  
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS refresh_sessions;
-- +goose StatementEnd
