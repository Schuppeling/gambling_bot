package database

/*List of queries that I use*/

const (
	//SELECTS
	GET_POINTS_BY_DISCORD_ID                 = `SELECT points from users where discord_id = $1;`
	GET_USERS_POINTS_DESC                    = `SELECT discord_id, points from users order by points desc;`
	GET_BET_ID_IF_EXISTS                     = `SELECT id from bets where exists(select 1 from bets where bet_name = $1 and active = true);`
	GET_BET_NAME_AND_POT                     = `SELECT bet_name, pot from bets where active = true;`
	GET_POT_BY_BET_NAME                      = `SELECT pot from bets where bet_name = $1;`
	GET_BIDDING_AMOUNT_BY_DISCORD_AND_BET_ID = `SELECT amount from bids where user_id = (select id from users where discord_id = $1) and bet_id = $2;`
	GET_BET_ID_AND_BID_AMOUNT_BY_BET_NAME    = `SELECT b.id, g.amount from bids g join bets b on g.bet_id = b.id where b.bet_name = $1;`
	GET_USER_IDS_AND_POINTS_BY_BET_ID        = `SELECT b.user_id, u.points from bids b join users u on u.id = b.user_id where b.bet_id = $1;`
	GET_BET_ID_AND_POT_BY_BET_NAME           = `SELECT id, pot from bets where bet_name = $1 and active = TRUE;`
	GET_BID_ID_FOR_BET_NAME                  = `SELECT id from bids where exists(SELECT 1 from bids where bet_id = (select id from bets where bet_name = $1 and active = TRUE) and user_id = (SELECT id from users where discord_id = $2));`

	//INSERTS
	INSERT_NEW_USER = `INSERT INTO users (discord_id, points) VALUES ($1, $2);`
	INSERT_NEW_BET  = `INSERT INTO bets (duration, bet_name, pot, active, result) VALUES ($1, $2, $3, $4, $5);`
	INSERT_NEW_BID  = `INSERT into bids (user_id, amount, bet_id, choice) values ((select id from users where discord_id = $1), $2, $3, true);`

	//UPDATES
	UPDATE_USERS_POINTS_BY_DISCORD_ID  = `UPDATE users set points = $1 where discord_id = $2;`
	UPDATE_BID_FOR_DISCORD_ID          = `UPDATE bids set amount = $1 where user_id = (select id from users where discord_id = $2);`
	UPDATE_POT_BY_BET_ID               = `UPDATE bets set pot = $1 where id = $2;`
	UPDATE_USERS_POINTS_BY_ID          = `UPDATE users set points = $1 where id = $2;`
	UPDATE_BET_TO_INACTIVE_WITH_RESULT = `UPDATE bets set active = FALSE, result = $1 where bet_name = $2 and active = TRUE;`
)
