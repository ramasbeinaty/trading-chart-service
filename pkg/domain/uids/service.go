package uids

import (
	"github.com/ramasbeinaty/trading-chart-service/pkg/domain/base/logger"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/clients/snowflake"
)

type UIDService struct {
	lgr             logger.ILogger
	snowflakeClient *snowflake.SnowflakeClient

	isDevMode bool
	idCounter int64
}

func NewUIDService(
	lgr logger.ILogger,
	isDevMode bool,
	snowflakeClient *snowflake.SnowflakeClient,
) *UIDService {
	return &UIDService{
		lgr:             lgr,
		isDevMode:       isDevMode,
		idCounter:       0,
		snowflakeClient: snowflakeClient,
	}
}

func (s *UIDService) GenerateUID() (int64, error) {
	if s.isDevMode {
		s.idCounter++
		return int64(s.idCounter), nil
	}

	return s.snowflakeClient.GenerateID()
}
