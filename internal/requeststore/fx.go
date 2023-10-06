package requeststore

import "go.uber.org/fx"

var Module = fx.Module("requests", 
        fx.Provide(NewRequestRepository),
        fx.Invoke(MigrateRequestRepository),
)
