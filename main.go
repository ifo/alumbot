package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ifo/gozulipbot"
)

func main() {
	bot := gozulipbot.Bot{}
	err := bot.GetConfigFromFlags()
	if err != nil {
		log.Fatalln(err)
	}
	bot.Init()

	// Only listen for private messages
	queue, err := bot.RegisterPrivate()
	if err != nil {
		log.Fatalln(err)
	}

	// Don't register the stop func because there's no intent to stop
	messages, _ := queue.EventsChan()

	stream := "alumni-checkins"

	for m := range messages {
		// Ignore the messages we see that are from us.
		if m.SenderEmail == bot.Email {
			continue
		}

		// send message to stream if we haven't already
		topic := getTodaysTopic()

		// If this is the message we intend to send, don't respond to it.
		// This is a second check like the above check to ensure we don't
		// respond indefinitely.
		if m.Content == makeTopicLocationMessage(stream, topic) {
			continue
		}

		// If the topic is new, send the message and create the stream
		// otherwise we've already done this so just respond to the person
		if topic != currentTopicCache {
			// Send the message to the new stream topic
			newMessage := gozulipbot.Message{
				Stream:  stream,
				Topic:   topic,
				Content: "Welcome to daily [checkins](https://github.com/ifo/alum-bot)!",
			}
			_, err := bot.Message(newMessage)
			if err != nil {
				m.Queue.Bot.Respond(m, "There was an error, sorry. The topic string should be: "+topic)
				// Ensure the cache is updated so we can't infinitely hit this code path
				currentTopicCache = topic
				continue
			}
			// Update the cache to today
			currentTopicCache = topic
		}

		// Respond to the DM
		responseMessage := makeTopicLocationMessage(stream, topic)
		m.Queue.Bot.Respond(m, responseMessage)
	}
}

const dateFormat = "Monday. January 02, 2006"

var currentTopicCache = ""

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
	topicUrl := zulipTopicUrlFormatting(topic)
	url := fmt.Sprintf("https://recurse.zulipchat.com/#narrow/stream/%s/subject/%s", stream, topicUrl)
	message := fmt.Sprintf("[The topic is here](%s).", url)
	return message
}

func zulipTopicUrlFormatting(topic string) string {
	// Periods must be done first, otherwise we'll start replacing our replacements
	periods := strings.Replace(topic, ".", ".2E", -1)
	spaces := strings.Replace(periods, " ", ".20", -1)
	commas := strings.Replace(spaces, ",", ".2C", -1)
	return commas
}
