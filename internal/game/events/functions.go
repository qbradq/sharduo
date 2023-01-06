package events

import (
	"github.com/qbradq/sharduo/internal/game"
	"github.com/qbradq/sharduo/lib/util"
)

var fnreg *util.Registry[string, func(game.Object, game.Object)] = util.NewRegistry[string, func(game.Object, game.Object)]("events")
