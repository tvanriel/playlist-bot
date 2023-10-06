package commands

import (
	"strconv"

	executor "github.com/mitaka8/playlist-bot/internal/command-executor"
	"github.com/mitaka8/playlist-bot/internal/requeststore"
	"go.uber.org/fx"
)

type NewRequestCommandParams struct {
        fx.In

        Requests *requeststore.RequestRepository
}

func NewRequestCommand(p NewRequestCommandParams) *RequestCommand {
        return &RequestCommand{
                Requests: p.Requests,
        }
}

type RequestCommand struct {
        Requests *requeststore.RequestRepository
}

var _ executor.Command = &RequestCommand{}

func (r *RequestCommand) Name() string {
        return "request"
}
func (r *RequestCommand) SkipsPrefix() bool {
        return false
}
func (r *RequestCommand) Apply(ctx *executor.Context) error {
        if len(ctx.Args) != 1 {
                ctx.Reply("Usage: request <id>")
                return nil
        }

        id, err := strconv.ParseUint(ctx.Args[0], 10, 32)
        if err != nil {
                return err
        }


        err = r.Requests.Append(ctx.Message.Author.ID, ctx.Message.GuildID, uint(id))
        if err != nil {
                return err
        }
        ctx.Reply("Appended")
        return nil
}
