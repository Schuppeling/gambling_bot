create table bids (
    id serial primary key,
    user_id integer not null,
    amount integer not null,
    bet_id integer not null,
    choice boolean not null,
    constraint fk_bets
        foreign key(bet_id)
            references bets(id),
    constraint fk_users
        foreign key(user_id)
            references users(id)
);