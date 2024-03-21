-- +goose Up
create table chat (
    id serial primary key,
    usernames text[] not null,
    created_at timestamp not null default now()
);

create table msg (
    id serial primary key,
    from_user text not null,
    msg_text text not null,
    created_at timestamp not null default now()
);

-- +goose Down
drop table chat;
drop table msg;