-- Table for storing users
create table users (
    id uuid primary key default uuidv7(),
    email text not null unique,
    password_hash bytea,
    email_verified boolean default false not null,
    profile_picture text,
    google_sub text unique,
    created_at timestamptz(3) not null default now(),
    next_monthly_grant_at timestamptz(3) not null
);
create unique index idx_users_google_sub on users(google_sub);
create unique index idx_users_email on users(email);

-- Table for managing user sessions via refresh tokens
create table refresh_sessions (
    id uuid primary key default uuidv7(),
    user_id uuid not null references users(id) on delete cascade,
    token_hash bytea not null unique,
    user_agent text,
    ip inet,
    created_at timestamptz(3) not null default now(),
    expires_at timestamptz(3) not null,
    revoked_at timestamptz(3)
);
create unique index idx_refresh_token_hash on refresh_sessions(token_hash);

-- Table for managing user credit grants
create type credit_grant_type as enum('monthly', 'topup');
create table credit_grants (
    id uuid primary key default uuidv7(),
    user_id uuid not null references users(id) on delete cascade,
    amount integer not null constraint amount_positive check (amount >= 0),
    remaining_amount integer not null constraint remaining_amount_positive check (remaining_amount >= 0),
    type credit_grant_type not null,
    expires_at timestamptz(3),
    created_at timestamptz(3) not null default now()
);
create index idx_credit_grants_active on credit_grants(user_id, expires_at) where remaining_amount > 0;

-- Table for managing user credit expenses
create table credit_expenses (
    id uuid primary key default uuidv7(),
    user_id uuid not null references users(id) on delete cascade,
    grant_id uuid not null references credit_grants(id) on delete cascade,
    spent integer not null,
    spent_at timestamptz(3) not null,
    created_at timestamptz(3) not null default now()
);
create index idx_credit_expenses_grant_id on credit_expenses(grant_id);

create table speeches (
    id uuid primary key default uuidv7(),
    user_id uuid not null references users(id) on delete cascade,
    in_lang text not null,
    out_lang text not null,
    started_at timestamptz(3) not null,
    ended_at timestamptz(3) not null
);

create table http_requests (
    id uuid primary key default uuidv7(),
    request_id uuid not null,
    method text not null,
    route text not null,
    path text not null,
    recieved_at timestamptz(3) not null,
    duration_ms integer not null,
    response_code smallint not null,
    request_headers_bytes integer not null,
    request_body_bytes integer not null,
    response_headers_bytes integer not null,
    response_body_bytes integer not null,
    ip inet not null,
    user_agent text,
    error text,
    created_at timestamptz(3) default now()
);
