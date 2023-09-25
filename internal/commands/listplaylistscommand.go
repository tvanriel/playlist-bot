package commands

import (
	executor "github.com/mitaka8/playlist-bot/internal/command-executor"
	"github.com/mitaka8/playlist-bot/internal/playliststore"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ListPlaylistCommand struct {
	log           *zap.Logger
	PlaylistStore *playliststore.PlaylistStore
}

type NewListPlaylistCommandParams struct {
	fx.In
	Logging       *zap.Logger
	PlaylistStore *playliststore.PlaylistStore
}

func NewListPlaylistCommand(p NewListPlaylistCommandParams) *ListPlaylistCommand {
	return &ListPlaylistCommand{
		log:           p.Logging,
		PlaylistStore: p.PlaylistStore,
	}
}

var _ executor.Command = &ListPlaylistCommand{}

func (c *ListPlaylistCommand) Name() string {
	return "list"
}

func (c *ListPlaylistCommand) SkipsPrefix() bool {
	return false
}

func (c *ListPlaylistCommand) Apply(ctx *executor.Context) error {
	if len(ctx.Args) != 1 {
		_, err := ctx.Reply("you must provide at least 1 (one) argument to this function.")
		return err

	}
	c.log.Info("Listing playlist",
		zap.String("guildId", ctx.Message.GuildID),
		zap.String("name", ctx.Args[0]),
	)
	playlists, err := c.PlaylistStore.ListPlaylists(ctx.Message.GuildID)
	if err != nil {
		return err
	}

	ctx.ReplyList(playlists)

	return nil

}
