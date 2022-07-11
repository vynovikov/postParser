package main

import (
	"postParser/internal/adapters/application"
	"postParser/internal/adapters/driven/grpc"
	"postParser/internal/adapters/driver"
	"postParser/internal/core"
)

func main() {
	transmitter := grpc.NewTransmitter()
	core := core.NewCore()
	app := application.NewApplication(core, transmitter)
	receiver := driver.NewReceiver(app)
	receiver.Run()
}
