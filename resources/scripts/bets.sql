create table bets (
    id      serial primary key,
	duration    integer not null,
	bet_name    varchar not null,
	pot     integer not null,
    active boolean not null,
    result boolean
);