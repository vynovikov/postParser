package main

import (
	"postParser/internal/adapters/application"
	"postParser/internal/adapters/driven/grpc"
	"postParser/internal/adapters/driven/store"
	"postParser/internal/adapters/driver/tp"
	"postParser/internal/adapters/driver/tps"
	"postParser/internal/core"
	"time"
)

func main() {
	transmitter := grpc.NewTransmitter()
	core := core.NewCore()
	store := store.NewStore()

	app := application.NewApplication(core, store, transmitter)

	tpReceiver := tp.NewTpReceiver(app)
	tpsReceiver := tps.NewTpsReceiver(app)
	go tpReceiver.Run()
	go tpsReceiver.Run()
	time.Sleep(time.Minute * 10)
}
