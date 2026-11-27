package main

import (
	"github.com/opencairn-org/cairn-core/internal/adapter/secondary/postgres"
)

type BasicApp struct {
	Activities *postgres.ActivityRepo
	Streams    *postgres.StreamRepo
}

func newBasicApp(runtime *appRuntime) (*BasicApp, error) {

	var basicApp BasicApp

	basicApp.Activities = postgres.NewActivityRepo(runtime.pool)
	basicApp.Streams = postgres.NewStreamRepo(runtime.pool)

	return &basicApp, nil
}
