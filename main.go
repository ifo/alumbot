package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ifo/gozulipbot"
)

const dateFormat = "Monday. January 02, 2006"

var currentTopicCache = ""

func main() {
	bot := gozulipbot.Bot{}
	err := bot.GetConfigFromFlags()
	if err != nil {
		log.Fatalln(err)
	}
	bot.Init()

	// We're only listening for private messages.
	queue, err := bot.RegisterPrivate()
	if err != nil {
		log.Fatalln(err)
	}

	// Don't register the stop func because we intend to run forever.
	queue.EventsCallback(startTopic)

	// Wait forever.
	stop := make(chan struct{})
	<-stop
}

func startTopic(em gozulipbot.EventMessage, err error) {
	if err != nil {
		log.Println(err)
		return
	}

	bot := em.Queue.Bot
	stream := "alumni-checkins"

	// Ignore the messages we see that are from us, otherwise we'll respond to ourselves.
	if em.SenderEmail == bot.Email {
		return
	}

	// If the message we received is the same message we intend to send, don't send it.
	// This is a second check like the one above to ensure we don't respond indefinitely.
	topic := getTodaysTopic()
	if em.Content == makeTopicLocationMessage(stream, topic) {
		return
	}

	// TODO: Based on github issue #1, update this to use an api call to see if a topic exists.
	// That will likely require changes to the bot library, gozulipbot.

	// If the topic is new, send the message and create the stream.
	// Otherwise we've already done this so just respond to the person.
	if topic != currentTopicCache {
		// Send the message to the new stream topic.
		newMessage := gozulipbot.Message{
			Stream:  stream,
			Topic:   topic,
			Content: "Welcome to daily [checkins](https://github.com/ifo/alum-bot)!",
		}
		_, err := bot.Message(newMessage)
		// Always update the topic cache to enable our second don't-infinitely-respond check above.
		currentTopicCache = topic
		if err != nil {

			bot.Respond(em, "There was an error, sorry. The topic string should be: "+topic)
			// Ensure the cache is updated so we can't infinitely hit this code path.
			currentTopicCache = topic
			return
		}
	}

	// Respond to the DM.
	responseMessage := makeTopicLocationMessage(stream, topic)
	bot.Respond(em, responseMessage)
}

func getTodaysTopic() string {
	// After all this work, it turns out that 16:00 Australia is UTC - 4
	// Which is probably an okay time for folks to be checking in.
	// So we'll just use that for now, and eventually figure out how to know
	// where people are, and use their current time.
	date := time.Now().In(time.FixedZone("UTC-4", -14400)).Format(dateFormat)
	topic := "Checkins! " + date
	return topic
}

func makeTopicLocationMessage(stream, topic string) string {
	topicURL := zulipTopicURLFormatting(topic)
	url := fmt.Sprintf("https://recurse.zulipchat.com/#narrow/stream/%s/subject/%s", stream, topicURL)
	message := fmt.Sprintf("[The topic is here](%s).", url)
	return message
}

func zulipTopicURLFormatting(topic string) string {
	// Periods must be replaced first, otherwise we'll start replacing our replacements.
	periods := strings.Replace(topic, ".", ".2E", -1)
	spaces := strings.Replace(periods, " ", ".20", -1)
	commas := strings.Replace(spaces, ",", ".2C", -1)
	return commas
}
