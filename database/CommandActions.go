package database

import (
	"database/sql"
	"fmt"
	"gambling/types"

	_ "github.com/lib/pq"
)

var Db *sql.DB

// func ConnectToDB(properties *types.DBProperties) error {
// 	var err error

// 	info := fmt.Sprintf("host=%s port=%d user=%s "+
// 		"password=%s dbname=%s sslmode=disable",
// 		properties.Host, properties.Port, properties.User, properties.Password, properties.Name)

// 	Db, err = sql.Open("postgres", info)

// 	if err != nil {
// 		return err
// 	}

// 	err = Db.Ping()

// 	return err
// }

/*
Gets the current amount of points belonging to a discord user. If the user does not exist, a row is
inserted.
*/
func GetPoints(discord_id string) int64 {
	num_points := getPointsByDiscordId(discord_id)

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

	updateUsersPointsByDiscordId(user, newPoints)
}

func GetRankings() ([]types.Rank, error) {
	var rank []types.Rank

	rows, err := Db.Query(GET_USERS_POINTS_DESC)
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
	rows, err := Db.Query(GET_BET_ID_IF_EXISTS, betName)
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

	_, err := Db.Exec(INSERT_NEW_BET, bet.Duration, bet.Name, bet.Pot, true, nil)
	if err != nil {
		fmt.Println(err)
	}

	return
}

func GetAllCurrentBets() ([]types.Bet, error) {
	var bets []types.Bet

	rows, err := Db.Query(GET_BET_NAME_AND_POT)
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

	rows, err := Db.Query(GET_POT_BY_BET_NAME, betName)
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

	rows, err := Db.Query(GET_BIDDING_AMOUNT_BY_DISCORD_AND_BET_ID, discordId, betId)
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
		rows, err = Db.Query(GET_BET_ID_AND_BID_AMOUNT_BY_BET_NAME, betName)

		for rows.Next() {
			err = rows.Scan(&betId, &gamblerAmount)
			if err != nil {
				return
			}
		}

		rows.Close()
		//update bids table
		var newGambleAmount int64 = gamblerAmount + amount
		_, err = Db.Exec(UPDATE_BID_FOR_DISCORD_ID, newGambleAmount, user)
		if err != nil {
			return
		}
	} else {
		rows, err = Db.Query(GET_BET_ID_IF_EXISTS, betName)

		for rows.Next() {
			err = rows.Scan(&betId)
			if err != nil {
				return
			}
		}

		rows.Close()

		var newGambleAmount int64 = gamblerAmount + amount
		_, err = Db.Exec(INSERT_NEW_BID, user, newGambleAmount, betId)
		if err != nil {
			return
		}
	}

	//update pot amount in bets table
	newPotAmount := potAmount + amount
	_, err = Db.Exec(UPDATE_POT_BY_BET_ID, newPotAmount, betId)
	if err != nil {
		return
	}

	//subtract amount from user's points
	rows, err = Db.Query(GET_POINTS_BY_DISCORD_ID, user)

	var currentAmount int64
	for rows.Next() {
		err = rows.Scan(&currentAmount)
		if err != nil {
			return
		}
	}

	rows.Close()
	newUserAmount := currentAmount - amount
	_, err = Db.Exec(UPDATE_USERS_POINTS_BY_DISCORD_ID, newUserAmount, user)
	if err != nil {
		return
	}
}

func EndBet(betName string, winner bool) {
	//get bet id and pot for betName
	rows, err := Db.Query(GET_BET_ID_AND_POT_BY_BET_NAME, betName)

	var betId int
	var pot int64
	for rows.Next() {
		err = rows.Scan(&betId, &pot)
		if err != nil {
			return
		}
	}
	rows.Close()

	rows, err = Db.Query(GET_USER_IDS_AND_POINTS_BY_BET_ID, betId)

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
	for k, v := range info {
		_, err := Db.Exec(UPDATE_USERS_POINTS_BY_ID, v+newAmount, k)
		if err != nil {
			return
		}
	}

	//update bets table active flag
	_, err = Db.Exec(UPDATE_BET_TO_INACTIVE_WITH_RESULT, winner, betName)
	if err != nil {
		return
	}
}

func bidExists(user string, betName string) bool {
	rows, err := Db.Query(GET_BID_ID_FOR_BET_NAME, betName, user)
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
