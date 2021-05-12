package database

import (
	"database/sql"
	"fmt"
	"gambling/types"

	_ "github.com/lib/pq"
)

var Db *sql.DB

func ConnectToDB(properties *types.DBProperties) error {
	var err error

	info := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		properties.Host, properties.Port, properties.User, properties.Password, properties.Name)

	Db, err = sql.Open("postgres", info)

	if err != nil {
		return err
	}

	err = Db.Ping()

	return err
}

/*
Gets the current amount of points belonging to a discord user. If the user does not exist, a row is
inserted.
*/
func GetPoints(discord_id string) int64 {
	var num_points int64 = 0
	sqlStatement := `SELECT points from users where discord_id = $1;`

	row := Db.QueryRow(sqlStatement, discord_id)

	switch err := row.Scan(&num_points); err {
	case sql.ErrNoRows:
		insertStatement := `INSERT INTO users (discord_id, points) VALUES ($1, $2);`

		_, err = Db.Exec(insertStatement, discord_id, num_points)
		if err != nil {
			panic(err)
		}

		break
	case nil:
		break
	default:
		panic(err)
	}

	return num_points
}

/*
Adds points to the specified user.
*/
func AddPoints(user string, points int64) {
	//get points for current user
	currentPoints := GetPoints(user)

	//add points and update record
	newPoints := currentPoints + points
	updateStatement := `UPDATE users set points = $1 where discord_id = $2;`

	_, err := Db.Exec(updateStatement, newPoints, user)
	if err != nil {
		panic(err)
	}
}

func GetRankings() ([]types.Rank, error) {
	var rank []types.Rank
	sqlStatement := `SELECT discord_id, points from users order by points desc;`

	rows, err := Db.Query(sqlStatement)
	defer rows.Close()

	for rows.Next() {
		var discord_id string
		var num_points int64

		err = rows.Scan(&discord_id, &num_points)
		if err != nil {
			return nil, err
		}

		r := types.Rank{
			User:   discord_id,
			Points: num_points,
		}

		rank = append(rank, r)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return rank, err
}

func betExists(betName string) bool {
	sqlStatement := `SELECT id from bets where exists(select 1 from bets where bet_name = $1 and active = true);`

	rows, err := Db.Query(sqlStatement, betName)
	defer rows.Close()

	if err != nil {
		fmt.Printf("Error getting whether bet=%s exists", betName)
		return true
	}

	if rows.Next() {
		return true
	}

	return false
}

//TODO create go routine to start the bet after creation
func CreateBet(bet *types.Bet) {
	exists := betExists(bet.Name)

	if exists {
		fmt.Println("Bet already exists")
		return
	}

	insertStatement := `INSERT INTO bets (duration, bet_name, pot, active, result) VALUES ($1, $2, $3, $4, $5);`

	_, err := Db.Exec(insertStatement, bet.Duration, bet.Name, bet.Pot, true, nil)
	if err != nil {
		fmt.Println(err)
	}

	return
}

func GetAllCurrentBets() ([]types.Bet, error) {
	var bets []types.Bet
	sqlStatement := `SELECT bet_name, pot from bets where active = true;`

	rows, err := Db.Query(sqlStatement)
	defer rows.Close()

	if err != nil {
		fmt.Printf("Error getting all bets")
		return nil, err
	}

	for rows.Next() {
		var betName string
		var pot int64

		err = rows.Scan(&betName, &pot)
		if err != nil {
			return nil, err
		}

		bet := types.Bet{
			Name: betName,
			Pot:  pot,
		}

		bets = append(bets, bet)
	}

	return bets, err
}

func GetPotAmount(betName string) (int64, error) {
	var amount int64 = 0
	sqlStatement := `SELECT pot from bets where bet_name = $1;`

	rows, err := Db.Query(sqlStatement, betName)
	defer rows.Close()

	if err != nil {
		fmt.Println("Error getting current pot amount")
		return amount, err
	}

	for rows.Next() {
		err = rows.Scan(&amount)
		if err != nil {
			return amount, err
		}
	}

	return amount, err
}

func getGambleAmount(discordId string, betId int) (int64, error) {
	var amount int64 = 0
	sqlStatement := `SELECT amount from bids where user_id = (select id from users where discord_id = $1) and bet_id = $2;`

	rows, err := Db.Query(sqlStatement, discordId, betId)
	defer rows.Close()

	if err != nil {
		fmt.Println("Error getting bid amount for bet")
		return amount, err
	}

	for rows.Next() {
		err = rows.Scan(&amount)
		if err != nil {
			return amount, err
		}
	}

	return amount, err
}

func AddBetAmount(user string, amount int64, betName string) {
	var gamblerAmount int64
	var rows *sql.Rows
	var err error
	var betId int

	potAmount, err := GetPotAmount(betName)
	if err != nil {
		fmt.Println(err)
	}

	if bidExists(user, betName) {
		queryStatement := `select b.id, g.amount from bids g join bets b on g.bet_id = b.id where b.bet_name = $1;`

		rows, err = Db.Query(queryStatement, betName)

		for rows.Next() {
			err = rows.Scan(&betId, &gamblerAmount)
			if err != nil {
				return
			}
		}

		rows.Close()
		//update bids table
		updateBids := `UPDATE bids set amount = $1 where user_id = (select id from users where discord_id = $2);`
		var newGambleAmount int64 = gamblerAmount + amount
		_, err = Db.Exec(updateBids, newGambleAmount, user)
		if err != nil {
			return
		}
	} else {
		queryStatement := `select id from bets b where b.bet_name = $1;`

		rows, err = Db.Query(queryStatement, betName)

		for rows.Next() {
			err = rows.Scan(&betId)
			if err != nil {
				return
			}
		}

		rows.Close()

		insertBids := `INSERT into bids (user_id, amount, bet_id, choice) values ((select id from users where discord_id = $1), $2, $3, true);`
		var newGambleAmount int64 = gamblerAmount + amount
		_, err = Db.Exec(insertBids, user, newGambleAmount, betId)
		if err != nil {
			return
		}
	}

	//update pot amount in bets table
	updateBets := `UPDATE bets set pot = $1 where id = $2;`
	newPotAmount := potAmount + amount
	_, err = Db.Exec(updateBets, newPotAmount, betId)
	if err != nil {
		return
	}

	//subtract amount from user's points
	getUserAmount := `select points from users where discord_id = $1;`
	rows, err = Db.Query(getUserAmount, user)

	var currentAmount int64
	for rows.Next() {
		err = rows.Scan(&currentAmount)
		if err != nil {
			return
		}
	}

	rows.Close()
	updateUsers := `UPDATE users set points = $1 where discord_id = $2;`
	newUserAmount := currentAmount - amount
	_, err = Db.Exec(updateUsers, newUserAmount, user)
	if err != nil {
		return
	}
}

func EndBet(betName string, winner bool) {
	//get bet id and pot for betName
	betIdStatement := `select id, pot from bets where bet_name = $1 and active = TRUE;`
	rows, err := Db.Query(betIdStatement, betName)

	var betId int
	var pot int64
	for rows.Next() {
		err = rows.Scan(&betId, &pot)
		if err != nil {
			return
		}
	}
	rows.Close()

	betterInfo := `select b.user_id, u.points from bids b join users u on u.id = b.user_id where b.bet_id = $1;`
	rows, err = Db.Query(betterInfo, betId)

	var count int64 = 0
	var user int
	var points int64
	info := make(map[int]int64)
	for rows.Next() {
		err = rows.Scan(&user, &points)
		if err != nil {
			return
		}
		info[user] = points
		count++
	}

	rows.Close()

	//update winners
	newAmount := pot / count
	updateWinner := `update users set points = $1 where id = $2;`
	for k, v := range info {
		_, err := Db.Exec(updateWinner, v+newAmount, k)
		if err != nil {
			return
		}
	}

	//update bets table active flag
	updateBets := `UPDATE bets set active = FALSE, result = $1 where bet_name = $2 and active = TRUE;`
	_, err = Db.Exec(updateBets, winner, betName)
	if err != nil {
		return
	}
}

func bidExists(user string, betName string) bool {
	sqlStatement := `SELECT id from bids where exists(select 1 from bids where bet_id = (select id from bets where bet_name = $1 and active = TRUE) and user_id = (select id from users where discord_id = $2));`

	rows, err := Db.Query(sqlStatement, betName, user)
	defer rows.Close()

	if err != nil {
		fmt.Printf("Error getting whether bid=%s exists", betName)
		return true
	}

	if rows.Next() {
		fmt.Println("Has rows")
		return true
	}

	return false
}
