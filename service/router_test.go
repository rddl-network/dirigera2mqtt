package service_test

import (
	"testing"

	"github.com/rddl-network/dirigera2mqtt/config"
	"github.com/rddl-network/dirigera2mqtt/service"

	"github.com/stretchr/testify/assert"
)

func TestTestnetModeTrue(t *testing.T) {
	cfg := config.DefaultConfig()

	s := service.NewTrustAnchorAttestationService(cfg)

	routes := s.GetRoutes()
	assert.Equal(t, 1, len(routes))
}
