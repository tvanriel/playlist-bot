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
		Log:             p.Log.Named("join"),
	}
}

func (j *JoinCommand) Name() string {
	return "join"
}

func (j *JoinCommand) SkipsPrefix() bool {
	return false
}

func (j *JoinCommand) Apply(ctx *executor.Context) error {

	log := j.Log.With(ctx.ZapFields()...)
	log.Info("Requested to Join Voicechannel")
	state, err := ctx.Session.State.VoiceState(ctx.Message.GuildID, ctx.Message.Author.ID)
	if err != nil {
		log.Warn("Error while requesting VoiceState data", zap.Error(err))
		return err
	}
	if state.ChannelID == "" {
		log.Info("Refusing to join voicechannel, user is not in a channel")
		ctx.Reply("You must be in a voicechannel to do this.")
		return nil
	}

	_, err = ctx.Session.ChannelVoiceJoin(ctx.Message.GuildID, state.ChannelID, false, true)

	if err != nil {
		log.Error("Cannot join voicechannel", zap.String("voiceChannelId", state.ChannelID), zap.Error(err))
		ctx.Reply(err.Error())
		return err
	}

	j.GuildRepository.JoinVoiceChannel(ctx.Message.GuildID, state.ChannelID)

	return nil
}
