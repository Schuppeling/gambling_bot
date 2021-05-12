package main

import (
	"fmt"
	"gambling/database"
	"gambling/handlers"
	"gambling/types"
	"gambling/utils"
	"log"
	"strconv"

	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"

	_ "github.com/lib/pq"
)

var dg *discordgo.Session

func init() {
	var err error
	utils.ReadProperties()

	host := utils.GetProperty("db.host")
	port, err := strconv.Atoi(utils.GetProperty("db.port"))
	user := utils.GetProperty("db.user")
	password := utils.GetProperty("db.password")
	dbName := utils.GetProperty("db.name")

	if err != nil {
		log.Fatal(err)
	}

	dbProperties := &types.DBProperties{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Name:     dbName,
	}

	err = database.ConnectToDB(dbProperties)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	var err error
	dg, err = discordgo.New("Bot " + utils.GetProperty("bot.token"))

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var err error

	//define goroutines
	go handlers.HydrationReminder(dg)

	//add handlers
	dg.AddHandler(handlers.MessageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
	//Close database connection
	database.Db.Close()
}
