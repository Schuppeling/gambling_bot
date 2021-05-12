create table users (
    id serial primary key,
    discord_id varchar not null,
    points integer not null
);