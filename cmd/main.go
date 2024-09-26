package main

import (
	"context"
	"log"
	"player/internal/config"
	"player/internal/rtmp"
	"player/internal/web"
	_ "player/logo"
)

func main() {

	ctx := context.Background()
	cfg := config.NewConfig()

	rtmpServer, err := rtmp.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}

	Group := web.NewGroup(ctx, rtmpServer, cfg)
	Group.Run(ctx)

}
