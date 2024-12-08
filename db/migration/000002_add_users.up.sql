CREATE TABLE "users" (
                         "id" bigserial PRIMARY KEY,
                         "username" varchar NOT NULL,
                         "hashed_password" varchar NOT NULL,
                         "full_name" varchar NOT NULL,
                         "email" varchar UNIQUE NOT NULL,
                         "password_changed_at" timestamptz NOT NULL DEFAULT('0001-01-01 00:00:00Z'),
                         "created_at" timestamptz NOT NULL DEFAULT (now())
);



ALTER TABLE "accounts" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "accounts" ADD CONSTRAINT "user_currency_key" UNIQUE ("user_id", "currency");