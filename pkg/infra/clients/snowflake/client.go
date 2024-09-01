package snowflake

import (
	"context"

	"github.com/bwmarrin/snowflake"
)

type SnowflakeClient struct {
	sfNode *snowflake.Node
}

func NewSnowflakeClient(
	ctx context.Context,
	config *SnowflakeConfig,
) *SnowflakeClient {
	sf, err := snowflake.NewNode(config.NodeNumber)
	if err != nil {
		panic(err)
	}

	return &SnowflakeClient{
		sfNode: sf,
	}
}

func (c *SnowflakeClient) GenerateID() (int64, error) {
	return int64(c.sfNode.Generate().Int64()), nil
}
