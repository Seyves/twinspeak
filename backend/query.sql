-- name: GetUser :one
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

-- name: CreateRefreshSession :one
insert into refresh_sessions (
    user_id, token_hash, user_agent, ip, expires_at
) values (
    $1, $2, $3, $4, $5
)
returning id;

-- name: GetRefreshSessionForUpdate :one
select * from refresh_sessions where token_hash = $1 for update;

-- name: RevokeRefreshSession :exec
update refresh_sessions set revoked_at = now() where token_hash = $1;
