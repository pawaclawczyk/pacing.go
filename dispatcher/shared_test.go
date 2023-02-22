package dispatcher

import (
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"testing"
)

func RunTestServer() *server.Server {
	return natsserver.RunServer(&natsserver.DefaultTestOptions)
}

func MakeTestConnection(t *testing.T) *nats.Conn {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("Cannot connect to NATS server: %v", err)
		return nil
	}
	return nc
}
