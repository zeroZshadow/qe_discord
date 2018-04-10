package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/kidoman/go-steam"
	"github.com/pkg/errors"
)

// Variables used for command line parameters
var (
	Token    string
	Address  string
	Server   *steam.Server
	LastInfo steam.InfoResponse
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&Address, "s", "", "Server ip")
	flag.Parse()

	if Token == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	var err error
	Server, err = connectSteamQuery(Address)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer Server.Close()

	dg, err := connectDiscordBot()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dg.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func connectSteamQuery(address string) (*steam.Server, error) {
	server, err := steam.Connect(address)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create steam connection")
	}
	return server, nil
}

func connectDiscordBot() (*discordgo.Session, error) {
	// Setup bot
	dg, err := discordgo.New("Bot " + Token)
	dg.ShouldReconnectOnError = true
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create Discord sessions")
	}

	// Register handlers
	dg.AddHandler(presenceUpdate)

	// Open connection to Discord
	err = dg.Open()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to open Discord connection")
	}
	return dg, nil
}

func presenceUpdate(s *discordgo.Session, m *discordgo.PresenceUpdate) {
	info, err := Server.Info()
	if err != nil {
		fmt.Println(err)
		return
	}

	if LastInfo.Players == info.Players {
		return
	}

	LastInfo = *info
	err = s.UpdateStatus(0, fmt.Sprintf("%d/%d Players online", info.Players, info.MaxPlayers))
	if err != nil {
		fmt.Println(err.Error())
	}
}
