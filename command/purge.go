package command

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	purgeUsage = "purge <amount> <@user>"
	purgeHelp  = "deletes <amount> messages sent by <user> in the current channel. Doesn't delete messages older than 14 days."
)

func purge(ctx *Context, args []string) {
	if len(args) < 3 {
		ctx.Reply("Usage: " + purgeUsage)
		return
	}

	number, err := strconv.Atoi(args[1])
	if err != nil {
		ctx.ReportError("the first argument must be a number", err)
		return
	} else if number > 100 || number < 2 {
		ctx.Reply("the first argument must be comprised between 2 and 100")
		return
	}

	from := parseMention(args[2])
	if from == "" {
		ctx.Reply("the second argument must be a user mention")
		return
	}

	var (
		toDelete = make([]string, 0, number)
		before   = ctx.Message.ID
		// discord doesn't let you bulk delete messages older than 14 days
		tooOldThreshold = (time.Hour * 24 * 14) - time.Hour
		now             = time.Now()
	)

Outer:
	for i := 1; i < 10; i++ {
		messages, err := ctx.Session.ChannelMessages(ctx.Message.ChannelID, 100, before, "", "")
		if err != nil {
			ctx.ReportError(fmt.Sprintf("could not fetch the 100 messages preceding message of id %s (likely missing permissions to read channel history)", before), err)
			if len(toDelete) > 0 {
				break Outer
			}
			return
		}
		for _, message := range messages {
			if created, _ := discordgo.SnowflakeTimestamp(message.ID); now.Sub(created) > tooOldThreshold {
				break Outer
			}
			if message.Author.ID != from {
				continue
			}

			toDelete = append(toDelete, message.ID)
			if len(toDelete) >= number {
				break Outer
			}
		}
		before = messages[len(messages)-1].ID
		if len(messages) < 100 {
			break
		}
	}
	err = ctx.Session.ChannelMessagesBulkDelete(ctx.Message.ChannelID, toDelete)
	if err != nil {
		ctx.ReportError("could not bulk delete messages (likely missing permissions)", err)
		return
	}

	ctx.Reply(fmt.Sprintf("Deleted %d messages", len(toDelete)))
}
