package mysql

import (
	"context"
	"testing"

	"github.com/artie-labs/transfer/lib/config"

	"github.com/stretchr/testify/suite"
)

type MySQLTestSuite struct {
	suite.Suite
	*Debezium
	ctx context.Context
}

func (m *MySQLTestSuite) SetupTest() {
	var debezium Debezium
	m.Debezium = &debezium
	m.ctx = context.Background()
	m.ctx = config.InjectSettingsIntoContext(m.ctx, &config.Settings{Config: &config.Config{}})
}

func TestPostgresTestSuite(t *testing.T) {
	suite.Run(t, new(MySQLTestSuite))
}
