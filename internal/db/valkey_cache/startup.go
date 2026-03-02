package valkey_cache

import (
	"context"
	"strconv"

	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/utils"
	glide "github.com/valkey-io/valkey-glide/go/v2"
	glideConf "github.com/valkey-io/valkey-glide/go/v2/config"
)

func GetDatabaseContext(ctx context.Context, config config.Config, logger utils.CustomJsonLogger, dbContext db.DbContext) *ValkeyCacheContext {
	port, err := strconv.Atoi(config.Cache.Port)
	if err != nil {
		logger.ErrorExit(ctx, err.Error())
	}

	valkeyConf := glideConf.NewClientConfiguration().WithAddress(&glideConf.NodeAddress{Host: config.Cache.Host, Port: port})

	client, err := glide.NewClient(valkeyConf)
	if err != nil {
		logger.ErrorExit(ctx, err.Error())
	}

	_, err = client.Ping(ctx)
	if err != nil {
		logger.ErrorExit(ctx, err.Error())
	}

	context := NewValkeyCacheContext(logger, client, dbContext)
	return context
}
