package commands

import (
	executor "github.com/mitaka8/playlist-bot/internal/command-executor"
	"github.com/mitaka8/playlist-bot/internal/playliststore"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type AddPlaylistCommand struct {
	log           *zap.Logger
	PlaylistStore *playliststore.PlaylistStore
}

type NewAddPlaylistCommandParams struct {
	fx.In
	Logging       *zap.Logger
	PlaylistStore *playliststore.PlaylistStore
}

func NewAddPlaylistCommand(p NewAddPlaylistCommandParams) *AddPlaylistCommand {
	return &AddPlaylistCommand{
		log:           p.Logging,
		PlaylistStore: p.PlaylistStore,
	}
}

var _ executor.Command = &AddPlaylistCommand{}

func (c *AddPlaylistCommand) Name() string {
	return "add-playlist"
}

func (c *AddPlaylistCommand) SkipsPrefix() bool {
	return false
}

func (c *AddPlaylistCommand) Apply(ctx *executor.Context) error {
	if len(ctx.Args) != 1 {
		_, err := ctx.Reply("you must provide at least 1 (one) argument to this function.")
		return err

	}

	name := ctx.Args[0]
	guildId := ctx.Message.GuildID

	c.log.Info("Adding playlist",
		zap.String("guildId", guildId),
		zap.String("name", name),
	)

	exists, err := c.PlaylistStore.PlaylistExists(guildId, name)
	if err != nil {
		return err
	}
	if exists {
		ctx.Reply("Playlist already exists!")
		return nil
	}

	err = c.PlaylistStore.AddPlaylist(guildId, name)
	if err != nil {
		return err
	}
	ctx.Reply("Added playlist!")

	return nil

}
