-- users table
CREATE TABLE "users"(
    "id" bigserial PRIMARY KEY,
    "username" varchar(50) UNIQUE NOT NULL,
    "email" varchar(255) UNIQUE NOT NULL,
    "password_hash" varchar(255) NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT (now())
);
--tweets table
CREATE TABLE "tweets"(
    "id" bigserial PRIMARY KEY,
    "user_id" bigint NOT NULL,
    "content" text NOT NULL,
    "image_url" varchar(255),
    "created_at" timestamptz NOT NULL DEFAULT(now())
);
--sessions table
CREATE TABLE "sessions" (
    "id" bigserial PRIMARY KEY,
    "user_id" bigint NOT NULL,
    "token_hash" varchar(255) UNIQUE NOT NULL,
    "expires_at" timestamptz NOT NULL 
);
ALTER TABLE "tweets" ADD FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE;
ALTER TABLE "sessions" ADD FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE;