-- name: GetUser :one
select * from users where email = $1;

-- name: GetUserByID :one
select * from users where id = $1;

-- name: CreateUser :one
insert into users (email, password_hash) values ($1, $2)
returning id;

-- name: CreateAccountFromGoogle :one
insert into users (
    google_sub, email, email_verified, profile_picture
) values (
    $1, $2, true, $3
)
returning id;

-- name: FindAccountFromGoogle :one
select id from users where google_sub = $1;

-- name: CreateSubscription :one
insert into subscriptions (user_id, next_monthly_grant_at) values ($1, $2)
returning id;

-- name: GetExpiredSubscriptions :many
select user_id from subscriptions where next_monthly_grant_at >= $1;

-- name: UpdateSubscription :exec
update subscriptions
set next_monthly_grant_at = $2
where user_id = $1;

-- name: CreateRefreshSession :one
insert into refresh_sessions (
    user_id, token_hash, user_agent, ip, expires_at
) values (
    $1, $2, $3, $4, $5
)
returning id;

-- name: GetRefreshSessionForUpdate :one
select * from refresh_sessions where token_hash = $1 for update;

-- name: GetRefreshSession :one
select * from refresh_sessions where token_hash = $1;

-- name: RevokeRefreshSession :exec
update refresh_sessions set revoked_at = now() where token_hash = $1;

-- name: CreateCreditGrant :exec
insert into credit_grants (
    user_id, amount, remaining_amount, type, expires_at
) values (
    $1, $2, $3, $4, $5
);

-- name: FindCreditGrantForSpend :one
select * from credit_grants 
where 
    user_id = $1 and 
    remaining_amount > 0 and
    (expires_at = null or expires_at > $2)
order by expires_at asc nulls last
limit 1 for update;

-- name: FindCreditGrantForSpend :one
select * from credit_grants 
where 
    user_id = $1 and 
    remaining_amount > 0 and
    (expires_at = null or expires_at > $2)
order by expires_at asc nulls last
limit 1 for update;

-- name: UpdateGrant :exec
update credit_grants 
set remaining_amount = $2
where id = $1;

-- name: CreateCreditExpenses :exec
insert into credit_expenses (
    user_id, grant_id, spent, spent_at
) values (
    $1, $2, $3, $4
);

-- name: InsertHttpRequest :exec
insert into http_requests (
    request_id,
    method,
    route,
    path,
    recieved_at,
    duration_ms,
    response_code,
    request_headers_bytes,
    request_body_bytes,
    response_headers_bytes,
    response_body_bytes,
    ip,
    user_agent,
    error
) values (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
);
create index idx_http_requests_received_at
on http_requests(received_at);

create index idx_http_requests_route
on http_requests(route);

create index idx_http_requests_response_code
on http_requests(response_code);

-- name: InsertSpeech :exec
insert into speeches (
    user_id, in_lang, out_lang, started_at, ended_at
) values (
    $1, $2, $3, $4, $5
);
