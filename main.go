package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	Token         string
	MainDirectory string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&MainDirectory, "i", "", "Image input path")
	flag.Parse()
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	var subDirectory = ""
	switch m.Content {
	case "&john", "&fru":
		subDirectory = "fru"
		break
	case "&tony":
		subDirectory = "tony"
		break
	case "&flea":
		subDirectory = "flea"
		break
	case "&chad":
		subDirectory = "chad"
		break
	case "&josh":
		subDirectory = "josh"
		break
	case "&cornell", "&chris":
		subDirectory = "chris"
		break
	default:
		return
	}
	imageDirectory := filepath.Join(MainDirectory, subDirectory)
	pick := pickRandomFile(imageDirectory)
	file, err := os.Open(filepath.Join(imageDirectory, pick.Name()))
	if err != nil {
		log.Panic(err)
	}
	discordFile := discordgo.File{
		Name:        pick.Name(),
		ContentType: "image/" + filepath.Ext(pick.Name()),
		Reader:      file,
	}
	var files = []*discordgo.File{&discordFile}
	data := discordgo.MessageSend{
		Files: files,
	}
	s.ChannelMessageSendComplex(
		m.ChannelID,
		&data,
	)
	return
}

var randSource = rand.NewSource(time.Now().UnixNano())
var rand1 = rand.New(randSource)

func pickRandomFile(dir string) os.DirEntry {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	var all []os.DirEntry
	for _, file := range files {
		fmt.Println(file.Name(), file.IsDir())
		if !file.IsDir() {
			all = append(all, file)
		}
	}
	randomIndex := rand1.Intn(len(all))
	pick := all[randomIndex]
	println(pick)
	return pick
}
