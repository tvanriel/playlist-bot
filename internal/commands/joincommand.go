package commands

import (
	executor "github.com/mitaka8/playlist-bot/internal/command-executor"
	"github.com/mitaka8/playlist-bot/internal/guildstore"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type JoinCommand struct {
	GuildRepository *guildstore.GuildRepository
	Log             *zap.Logger
}

type NewJoinCommandParams struct {
	fx.In
	Log             *zap.Logger
	GuildRepository *guildstore.GuildRepository
}

func NewJoinCommand(p NewJoinCommandParams) *JoinCommand {
	return &JoinCommand{
		GuildRepository: p.GuildRepository,
		Log:             p.Log,
	}
}

func (j *JoinCommand) Name() string {
	return "join"
}

func (j *JoinCommand) SkipsPrefix() bool {
	return false
}

func (j *JoinCommand) Apply(ctx *executor.Context) error {
	state, err := ctx.Session.State.VoiceState(ctx.Message.GuildID, ctx.Message.Author.ID)
	if err != nil {
		return err
	}
	if state.ChannelID == "" {
		ctx.Reply("You must be in a voicechannel to do this.")
		return nil
	}

	_, err = ctx.Session.ChannelVoiceJoin(ctx.Message.GuildID, state.ChannelID, false, true)

	if err != nil {
		ctx.Reply(err.Error())
		return err
	}

	j.GuildRepository.JoinVoiceChannel(ctx.Message.GuildID, state.ChannelID)

	return nil
}
