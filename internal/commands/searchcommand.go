package commands

import (
	"strings"

	executor "github.com/mitaka8/playlist-bot/internal/command-executor"
	"github.com/mitaka8/playlist-bot/internal/playliststore"
	"go.uber.org/fx"
)

type NewSearchCommandParams struct {
        fx.In

        PlaylistStore *playliststore.PlaylistStore
}

func NewSearchCommand(p NewSearchCommandParams) (*SearchCommand) {
        return &SearchCommand{
                PlaylistStore: p.PlaylistStore,
        }
}

type SearchCommand struct {
        PlaylistStore *playliststore.PlaylistStore
        
}

var _ executor.Command = &SearchCommand{}

func (s *SearchCommand) SkipsPrefix() bool{
        return false
}

func (s *SearchCommand) Name() string {
        return "search"
}

func (s *SearchCommand) Apply(context *executor.Context) error {

        if len(context.Args) != 1{ 
                context.Reply("Usage: search <term>")
                return nil
        }
        term := strings.Join(context.Args, " ")
        guildId := context.Message.GuildID

        results, err := s.PlaylistStore.Search(guildId, term)
        if err != nil {
                return err
        }

        if len(results) == 0 {
                context.Reply("I can't find anything like that.")
                return nil
        }

        list := make([]string, len(results))
        for i := range results {
                list[i] = results[i].String()
        }
        context.ReplyList(list)
        return nil
}
