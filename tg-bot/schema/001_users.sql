-- +goose Up
create table if not exists users (
    telegram_id bigint primary key,
    chat_id bigint not null,
    positive_score integer not null default 0,
    negative_score integer not null default 0,
    created_at text not null default (CURRENT_TIMESTAMP),
    updated_at text not null default (CURRENT_TIMESTAMP)
);

-- +goose Down
drop table if exists users;
