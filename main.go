package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/bwmarrin/discordgo"
	"github.com/h2non/bimg"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
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
	log.Println("Starting frubot")
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
		log.Fatal("error opening connection,", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func messageCreate(s *discordgo.Session, mc *discordgo.MessageCreate) {
	m := mc.Message
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "&center" {
		log.Println("Received center message")
		err := center(s, m)
		if err != nil {
			_, err := s.ChannelMessage(m.ChannelID, "Failed to center image")
			if err != nil {
				log.Println(err.Error())
			}
			log.Println(err.Error())
		}
		return
	}

	subDirectory := determineSubdirectory(m)
	if subDirectory != "" {
		log.Println("Received image request: " + subDirectory)
		err := sendImage(s, m, subDirectory)
		if err != nil {
			s.ChannelMessage(m.ChannelID, "Failed to send image")
		}
	}
	return
}

func determineSubdirectory(m *discordgo.Message) string {
	var subDirectory string
	switch m.Content {
	case "&john", "&fru":
		subDirectory = "fru"
	case "&tony":
		subDirectory = "tony"
	case "&flea":
		subDirectory = "flea"
	case "&chad":
		subDirectory = "chad"
	case "&josh":
		subDirectory = "josh"
	case "&cornell", "&chris":
		subDirectory = "chris"
	default:
		subDirectory = ""
	}
	return subDirectory
}

func sendImage(s *discordgo.Session, m *discordgo.Message, subDirectory string) error {
	imageDirectory := filepath.Join(MainDirectory, subDirectory)
	pick, err := pickRandomFile(imageDirectory)
	if err != nil {
		return err
	}
	file, err := os.Open(filepath.Join(imageDirectory, pick.Name()))
	if err != nil {
		return err
	}
	sendFile(s, m, pick.Name(), file)
	return nil
}

func sendFile(s *discordgo.Session, m *discordgo.Message, name string, file *os.File) {
	discordFile := discordgo.File{
		Name:        name,
		ContentType: "image/" + filepath.Ext(name),
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
}

func center(s *discordgo.Session, m *discordgo.Message) error {
	urls := grabUrlsOfMessage(m.ReferencedMessage)
	if len(urls) == 0 {
		log.Println("Found no urls. Scanning message history")
		msgs, err := s.ChannelMessages(m.ChannelID, 50, "", "", "")
		if err != nil {
			return err
		}
		for _, msg := range msgs {
			urls = grabUrlsOfMessage(msg)
			if len(urls) != 0 {
				log.Println("Found urls " + strings.Join(urls, ", "))
				break
			}
		}
		if len(urls) == 0 {
			log.Println("Found no suitable URL in message history.")
			return nil
		}
	} else {
		log.Println("Found urls " + strings.Join(urls, ", "))
	}
	file, err := os.CreateTemp(os.TempDir(), "*"+path.Base(urls[0]))
	if err != nil {
		return err
	}
	err = downloadFile(urls[0], file.Name())
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	readFile, err := os.ReadFile(file.Name())
	if err != nil {
		return err
	}
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	svc := rekognition.NewFromConfig(cfg)
	input := &rekognition.DetectFacesInput{
		Image: &types.Image{
			Bytes: readFile,
		},
	}
	result, err := svc.DetectFaces(ctx, input)
	if err != nil {
		return err
	}
	buffer, err := bimg.Read(file.Name())
	if err != nil {
		return err
	}
	image := bimg.NewImage(buffer)
	size, err := image.Size()
	if err != nil {
		return err
	}
	var originalHeight = size.Height
	var originalWidth = size.Width
	var firstNoseCoordinateX int
	var firstNoseCoordinateY int
	var newAnchorX int
	var newAnchorY int
	var newHeight int
	var newWidth int

	for _, d := range result.FaceDetails {
		for _, landmark := range d.Landmarks {
			if landmark.Type == "nose" {
				firstNoseCoordinateX = int(*landmark.X * float32(originalWidth))
				firstNoseCoordinateY = int(*landmark.Y * float32(originalHeight))
				newWidth = originalWidth - firstNoseCoordinateX
				if firstNoseCoordinateX*2 < originalWidth {
					newWidth = firstNoseCoordinateX * 2
					newAnchorX = 0
				} else {
					newWidth = (originalWidth - firstNoseCoordinateX) * 2
					newAnchorX = originalWidth - newWidth
				}
				if firstNoseCoordinateY*2 < originalHeight {
					newHeight = firstNoseCoordinateY * 2
					newAnchorY = 0
				} else {
					newHeight = (originalHeight - firstNoseCoordinateY) * 2
					newAnchorY = originalHeight - newHeight
				}
				break
			}
		}
	}
	resizedData, err := image.Extract(newAnchorY, newAnchorX, newWidth, newHeight)
	if err != nil {
		return err
	}
	err = os.WriteFile(file.Name(), resizedData, 0777)
	if err != nil {
		return err
	}
	sendFile(s, m, file.Name(), file)
	return nil
}

func grabUrlsOfMessage(m *discordgo.Message) []string {
	if m == nil {
		return []string{}
	}
	var urls []string
	for _, embed := range m.Embeds {
		urls = append(urls, embed.URL)
	}
	for _, embed := range m.Attachments {
		urls = append(urls, embed.URL)
	}
	return urls
}

var randSource = rand.NewSource(time.Now().UnixNano())
var rand1 = rand.New(randSource)

func pickRandomFile(dir string) (os.DirEntry, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var all []os.DirEntry
	for _, file := range files {
		if !file.IsDir() {
			all = append(all, file)
		}
	}
	randomIndex := rand1.Intn(len(all))
	pick := all[randomIndex]
	return pick, nil
}

func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
