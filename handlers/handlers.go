package handlers

import (
	"fmt"
	"gambling/database"
	"gambling/types"
	"gambling/utils"
	"time"

	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var commands []string = []string{
	"!bet <amount> <bet name>: Bet a specified amount on a specified bet",
	"!award <user> <point value>: Awards points to the specified user. (Can't be yourself)",
	"!points <user>: Says how many points a user has",
	"!rankings: Show the points leaderboard",
	"!createbet <bet name> <duration>: Create a bet with a bet name and specified duration",
	"!currentbets: Show the current running bets",
	"!pot <bet name>: Show current pot for a specified bet",
	"!end <bet name> <result>: End the bet early and specify the result"}

var textValidation []string = []string{
	`^!gambling$`,
	`^!award\s<@!\d+>\s\d+$`,
	`^!points\s<@!\d+>$`,
	`^!rankings$`,
	`^!bet\s\d+\s[\d\w]+$`,
	`^!createbet\s[\d\w]+\s\d+$`,
	`^!currentbets$`,
	`^!pot\s[\d\w]+$`,
	`^!end\s[\d\w]+\s[YN]$`}

func buildCommandsEmbedFields() []*discordgo.MessageEmbedField {
	var fields []*discordgo.MessageEmbedField

	for _, command := range commands {
		part := strings.Split(command, ": ")
		field := *&discordgo.MessageEmbedField{
			Name:  part[0],
			Value: part[1],
		}

		fields = append(fields, &field)
	}

	return fields
}

func buildRankingsEmbedFields(s *discordgo.Session) ([]*discordgo.MessageEmbedField, error) {
	var fields []*discordgo.MessageEmbedField
	var err error

	allRankings, err := database.GetRankings()
	if err != nil {
		return nil, err
	}
	for i, rank := range allRankings {
		var user *discordgo.User
		user, err = s.User(rank.User)

		if err != nil {
			fmt.Println(err)
			return fields, err
		}
		field := *&discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("%v. %v", i+1, user.Username),
			Value: strconv.FormatInt(rank.Points, 10),
		}

		fields = append(fields, &field)
	}

	return fields, err
}

func buildCurrentBetsEmbed(s *discordgo.Session, currentBets *[]types.Bet) []*discordgo.MessageEmbedField {
	var fields []*discordgo.MessageEmbedField

	for _, bet := range *currentBets {
		field := *&discordgo.MessageEmbedField{
			Name:  bet.Name,
			Value: fmt.Sprintf("Current Pot: %v", bet.Pot),
		}

		fields = append(fields, &field)
	}

	return fields
}

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	valid := isValidCommand(m.Content)
	if !valid {
		return
	}

	command := strings.Split(m.Content, " ")

	switch command[0] {
	case "!gambling":
		s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title:  "Commands",
			Color:  0x2abd42,
			Fields: buildCommandsEmbedFields(),
		})
		break
	case "!award":
		person := parsePerson(command[1])
		//To prevent adding points to yourself
		if person == m.Author.ID {
			s.ChannelMessageSend(m.ChannelID, "You can't add points to yourself!")
			break
		}

		//Add points if the amount is valid
		if points, err := strconv.ParseInt(command[2], 0, 64); err == nil {
			database.AddPoints(person, points)
		} else {
			fmt.Println(err)
		}
		break
	case "!points":
		person := parsePerson(command[1])
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@!%v> has %v points.", person, database.GetPoints(person)))
		break
	case "!rankings":
		fields, err := buildRankingsEmbedFields(s)

		if err != nil {
			break
		}

		s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title:  "Leaderboard",
			Color:  0xfae825,
			Fields: fields,
		})
		break
	case "!createbet": //<text> <duration>
		name := command[1]
		duration, err := strconv.Atoi(command[2])

		if err != nil {
			break
		}

		//create bet object to save
		bet := &types.Bet{
			Duration: duration,
			Name:     name,
			Pot:      0,
		}

		database.CreateBet(bet)
		break
	case "!currentbets":
		currentBets, err := database.GetAllCurrentBets()

		if err != nil {
			break
		}

		fields := buildCurrentBetsEmbed(s, &currentBets)

		s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
			Title:  "Current Bets",
			Color:  0xfae825,
			Fields: fields,
		})
		break
	case "!pot": //!pot <name of bet> to see current amount in pot
		pot, err := database.GetPotAmount(command[1])
		if err != nil {
			break
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Current pot for %v: %v", command[1], pot))
		break
	case "!end": //!end <name of bet> <winner?> to end bet early
		database.EndBet(command[1], command[2] == "Y")
		break
	case "!bet": //!bet <amount> <name of bet>
		amount, err := strconv.ParseInt(command[1], 0, 64)
		if err != nil {
			break
		}
		database.AddBetAmount(m.Author.ID, amount, command[2])
		break
	default:
		fmt.Println("Invalid Command")
	}
}

func HydrationReminder(s *discordgo.Session) {
	for range time.Tick(time.Hour * 1) {
		s.ChannelMessageSend(utils.GetProperty("bot.dm_channel"), "@here Drink some water!")
	}
}

//When you @ a person, the id comes in as <@!....>
func parsePerson(person string) string {
	return person[3 : len(person)-1]
}

func isValidCommand(input string) bool {
	for _, pattern := range textValidation {
		valid, err := regexp.MatchString(pattern, input)
		if err != nil {
			fmt.Print(err)
			break
		}

		if valid {
			return true
		}
	}

	return false
}
