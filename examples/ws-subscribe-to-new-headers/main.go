package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/centrifugal/centrifuge-go"
)

func main() {

	client := centrifuge.NewJsonClient("ws://localhost:8080/connection/websocket", centrifuge.Config{
		// Uncomment and adjust value if block headers service has authentication turned on
		Token: "mQZQ6WmxURxWz5ch",
	})
	defer client.Close()

	// It's only for example purposes, configure it according to your needs
	configureAdditionalDebugMessagesOnClient(client)

	err := client.Connect()
	if err != nil {
		log.Println(err)
	}

	sub, err := client.NewSubscription("headers", centrifuge.SubscriptionConfig{
		Recoverable: true,
		Positioned:  true,
	})

	// It's only for example purposes, configure it according to your needs
	configureAdditionalDebugMessagesOnSubscription(sub)

	sub.OnPublication(func(e centrifuge.PublicationEvent) {
		//TODO: HERE PLACE THE LOGIC TO HANDLE EVENT ABOUT NEW HEADER
		log.Printf("Event received on channel %s: %s (offset %d)", sub.Channel, e.Data, e.Offset)
	})

	err = sub.Subscribe()
	if err != nil {
		log.Println(err)
	}

	// Run some example of console to give some possibility to play around with websocket connection to block headers service
	quit := make(chan bool, 1)

	go handleUserInput(client, sub, quit)

	<-quit
	log.Println("Quitting")
}

func configureAdditionalDebugMessagesOnClient(client *centrifuge.Client) {
	client.OnConnecting(func(e centrifuge.ConnectingEvent) {
		log.Printf("[Client] Connecting - Status: %d (%s)", e.Code, e.Reason)
	})

	client.OnConnected(func(e centrifuge.ConnectedEvent) {
		log.Printf("[Client] Connected with ID %s", e.ClientID)
	})

	client.OnDisconnected(func(e centrifuge.DisconnectedEvent) {
		log.Printf("[Client] Disconnected - Status: %d (%s)", e.Code, e.Reason)
	})

	client.OnError(func(e centrifuge.ErrorEvent) {
		log.Printf("[Client] Error: %s", e.Error.Error())
	})
}

func configureAdditionalDebugMessagesOnSubscription(sub *centrifuge.Subscription) {
	sub.OnSubscribing(func(e centrifuge.SubscribingEvent) {
		log.Printf("[Subscription] Subscribing on channel %s - Status: %d (%s)", sub.Channel, e.Code, e.Reason)
	})
	sub.OnSubscribed(func(e centrifuge.SubscribedEvent) {
		log.Printf("[Subscription] Subscribed on channel %s, (%v)", sub.Channel, e)
	})
	sub.OnUnsubscribed(func(e centrifuge.UnsubscribedEvent) {
		log.Printf("[Subscription] Unsubscribed from channel %s - Status: %d (%s)", sub.Channel, e.Code, e.Reason)
	})

	sub.OnError(func(e centrifuge.SubscriptionErrorEvent) {
		log.Printf("Subscription error %s: %s", sub.Channel, e.Error)
	})
}

func handleUserInput(client *centrifuge.Client, sub *centrifuge.Subscription, quit chan bool) {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		switch text {
		case "#status":
			{
				log.Printf("Client: %v", client.State())
				log.Printf("Subscription: %v", sub.State())
			}
		case "#subscribe":
			err := sub.Subscribe()
			if err != nil {
				log.Println(err)
			}
		case "#unsubscribe":
			err := sub.Unsubscribe()
			if err != nil {
				log.Println(err)
			}
		case "#disconnect":
			err := client.Disconnect()
			if err != nil {
				log.Println(err)
			}
		case "#connect":
			err := client.Connect()
			if err != nil {
				log.Println(err)
			}
		case "#close":
			client.Close()
			quit <- true
		default:
			log.Print(
				`Unknown command` + text + `
Use following commands:
	#status - show status of connection and subscription
	#subscribe - subscribe to new header events
	#unsubscribe - unsubscribe to new header events
	#connect - connect to block headers service via websocket
	#disconnect - connect from block headers service via websocket
	#close - close example

`)
		}
	}
}
