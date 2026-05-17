create table users (
    id uuid primary key default gen_random_uuid(),
    email text not null unique,
    password_hash bytea,
    email_verified boolean default false not null,
    profile_picture text,
    google_sub text unique,
    created_at timestamptz(3) not null default now()
);

create unique index idx_users_google_sub on users(google_sub);
create unique index idx_users_email on users(email);

create table refresh_sessions (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references users(id),
    token_hash bytea not null unique,
    user_agent text,
    ip inet,
    created_at timestamptz(3) not null default now(),
    expires_at timestamptz(3) not null,
    revoked_at timestamptz(3)
);

create unique index idx_refresh_token_hash on refresh_sessions(token_hash);
