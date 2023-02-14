package dispatcher

import (
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
)

func RunServer() *server.Server {
	return natsserver.RunServer(&natsserver.DefaultTestOptions)
}
