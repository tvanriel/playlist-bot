package executor

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type Command interface {
	Name() string
	Apply(*Context) error
	SkipsPrefix() bool
}

type Context struct {
	Message *discordgo.Message
	Content string
	Args    []string
	Session *discordgo.Session
}

func (c *Context) ZapFields() []zap.Field {
	return []zap.Field{
		zap.Strings("args", c.Args),
		zap.String("guildId", c.Message.GuildID),
		zap.String("channelId", c.Message.ChannelID),
		zap.String("authorId", c.Message.Author.ID),
		zap.String("messageId", c.Message.Author.ID),
		zap.String("content", c.Content),
		zap.String("authorName", c.Message.Author.Username),
	}
}

func (ctx *Context) Reply(s string) (*discordgo.Message, error) {
	return ctx.Session.ChannelMessageSendEmbedReply(
		ctx.Message.ChannelID,

		&discordgo.MessageEmbed{
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Result",
					Value: s,
				},
			},
		},
		ctx.Reference(),
	)

}

func (ctx *Context) Error(err error) (*discordgo.Message, error) {
	return ctx.Session.ChannelMessageSendEmbedReply(
		ctx.Message.ChannelID,
		&discordgo.MessageEmbed{
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Error",
					Value: err.Error(),
				},
			},
			Color: 0xFF0000,
		},
		ctx.Reference(),
	)
}

func (ctx *Context) Reference() *discordgo.MessageReference {

	return &discordgo.MessageReference{
		MessageID: ctx.Message.ID,
		ChannelID: ctx.Message.ChannelID,
		GuildID:   ctx.Message.GuildID,
	}
}

const DISCORD_MSG_MAX_LEN = 2000

func (ctx *Context) ReplyList(s []string) ([]*discordgo.Message, error) {
	if len(s) == 0 {
		msg, err := ctx.Session.ChannelMessageSendReply(

			ctx.Message.ChannelID,
			" - nothing here",
			ctx.Reference(),
		)
		return []*discordgo.Message{msg}, err

	}

	itemTpl := "%s\n"
	templated := make([]string, len(s))
	for i := range s {
		templated = append(templated, fmt.Sprintf(itemTpl, s[i]))
	}
	var sb strings.Builder
	totalLength := 0

	var sentMessages []*discordgo.Message
	var err error
	for i := range templated {
		if totalLength+len(templated[i]) > DISCORD_MSG_MAX_LEN {
			msg, msgerr := ctx.Session.ChannelMessageSendReply(
				ctx.Message.ChannelID,
				sb.String(),
				ctx.Reference(),
			)
			sentMessages = append(sentMessages, msg)
			err = multierr.Append(err, msgerr)
			sb.Reset()
			totalLength = 0
		}

		totalLength += len(templated[i])
		sb.WriteString(templated[i])
	}
	if totalLength > 0 {
		msg, msgerr := ctx.Session.ChannelMessageSendReply(
			ctx.Message.ChannelID,
			sb.String(),
			ctx.Reference(),
		)
		sentMessages = append(sentMessages, msg)
		err = multierr.Append(err, msgerr)
	}
	return sentMessages, err

}
