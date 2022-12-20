package main

import (
	"postParser/internal/adapters/application"
	"postParser/internal/adapters/driven/grpc"
	"postParser/internal/adapters/driven/store"
	"postParser/internal/adapters/driver"
	"postParser/internal/core"
)

func main() {
	transmitter := grpc.NewTransmitter()
	core := core.NewCore()
	store := store.NewStore()

	app := application.NewApplication(core, store, transmitter)

	receiver := driver.NewReceiver(app)
	receiver.Run()
}
