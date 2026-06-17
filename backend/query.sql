-- name: GetUser :one
select * from users where email = $1;

-- name: GetUserByID :one
select * from users where id = $1;

-- name: GetUserByEmail :one
select * from users where email = $1;

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

-- name: LinkAccountToGoogle :one
update users set 
    profile_picture = $1, 
    google_sub = $2,
    email_verified = true
where email = $3
returning id;

-- name: InsertUserPrefs :exec
insert into preferences(user_id) values ($1);

-- name: GetUserPrefs :one
select * from preferences where user_id = $1;

-- name: UpdateUserPrefs :exec
update preferences
set 
    chat_message_size = $2,
    theme = $3,
    in_lang = $4,
    out_lang = $5,
    updated_at = now()
where user_id = $1;

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
order by expires_at asc nulls last, type asc, id asc
limit 1 for update;

-- name: GetUserCreditGrants :many
select id, user_id, amount, remaining_amount, type, expires_at, created_at from credit_grants
where
    user_id = $1 and
    (expires_at is null or expires_at > $2)
order by type asc, expires_at asc nulls last, id asc;

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
    user_id, in_lang, out_lang, transcription, translation, chat_side, started_at, ended_at
) values (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: GetSpeeches :many
select * from (
    select * from speeches 
    where user_id = $1 
    order by started_at desc 
    limit $2
) as s
order by s.started_at asc;

-- name: CreateVerificationToken :one
insert into email_verification_tokens (user_id, token_hash, expires_at)
values ($1, $2, $3)
returning id;

-- name: GetVerificationToken :one
select * from email_verification_tokens 
where token_hash = $1 and expires_at > now();

-- name: VerifyUserEmail :exec
update users set email_verified = true where id = $1;

-- name: DeleteVerificationToken :exec
delete from email_verification_tokens where token_hash = $1;

-- name: DeleteExpiredVerificationTokens :exec
delete from email_verification_tokens where expires_at <= now();
