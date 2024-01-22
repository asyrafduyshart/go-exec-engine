package pubnub

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	pubnub "github.com/pubnub/go"
)

// Data ...
type Data struct {
	PB *pubnub.PubNub
}

// Init ...
func Init() *Data {
	godotenv.Load()
	config := pubnub.NewConfig()
	fmt.Println(fmt.Sprintf("PUBNUB_SUBSCRIBE_KEY: %s", os.Getenv("PUBNUB_SUBSCRIBE_KEY")))
	config.SubscribeKey = os.Getenv("PUBNUB_SUBSCRIBE_KEY")
	config.PublishKey = os.Getenv("PUBNUB_PUBLISH_KEY")
	// config.UUID = os.Getenv("PUBNUB_UUID")
	// config.PublishKey = "pub-c-8defa790-667d-4e66-a855-cd293ee0b435"
	// config.SubscribeKey = "sub-c-78d4d302-727d-11eb-8178-92dbc0211330"

	pn := pubnub.NewPubNub(config)

	listener := pubnub.NewListener()

	waitForConnect := make(chan bool)

	go func() {
		for {
			select {
			case status := <-listener.Status:
				switch status.Category {
				case pubnub.PNConnectedCategory:
					// Connect event. You can do stuff like publish, and know you'll get it.
					// Or just use the connected event to confirm you are subscribed for
					// UI / internal notifications, etc
					waitForConnect <- true
				}
			case msg := <-listener.Message:
				fmt.Println(" --- MESSAGE: ")
				fmt.Println(fmt.Sprintf("msg.Channel: %s", msg.Channel))
				fmt.Println(fmt.Sprintf("msg.Message: %s", msg.Message))
				fmt.Println(fmt.Sprintf("msg.SubscribedChannel: %s", msg.SubscribedChannel))
				fmt.Println(fmt.Sprintf("msg.Timetoken: %d", msg.Timetoken))
			case presence := <-listener.Presence:
				fmt.Println(" --- PRESENCE: ")
				fmt.Println(fmt.Sprintf("%s", presence.Channel))
			}
		}
	}()

	pn.AddListener(listener)

	pn.Subscribe().
		Channels([]string{os.Getenv("PUBNUB_CHANNEL_LISTENER")}).
		WithPresence(true).
		Execute()

	<-waitForConnect
	fmt.Println("PUBNUB Connected")

	return &Data{
		PB: pn,
	}
}

// PushMessage ..
func (pnb *Data) PushMessage(channel, message string) (*pubnub.PublishResponse, pubnub.StatusResponse, error) {
	return pnb.PB.Publish().Channel(channel).Message(message).Execute()
}
