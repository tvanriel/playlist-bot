package executor

import "go.uber.org/fx"

const GROUP_COMMANDS = `group:"commands"`

var Module = fx.Module("command-executor",
	fx.Provide(
		fx.Annotate(
			NewCommandExecutor,
			fx.ParamTags(GROUP_COMMANDS),
		),
	),
)

func AsCommand(in any) any {
	return fx.Annotate(
		in,
		fx.As(new(Command)),
		fx.ResultTags(GROUP_COMMANDS),
	)
}
