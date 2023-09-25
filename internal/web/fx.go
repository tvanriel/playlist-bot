package web

import (
	"github.com/tvanriel/cloudsdk/http"
	"go.uber.org/fx"
)

var Module = fx.Module("web",
	fx.Decorate(DecorateTemplater),
	fx.Provide(
		NewTemplater,
		http.AsRouteGroup(NewWebInterface),
	),
)
