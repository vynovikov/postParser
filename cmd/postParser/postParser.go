// Central point
package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/vynovikov/postParser/internal/adapters/application"
	"github.com/vynovikov/postParser/internal/adapters/driven/rpc"
	"github.com/vynovikov/postParser/internal/adapters/driven/store"
	"github.com/vynovikov/postParser/internal/adapters/driver/tp"
	"github.com/vynovikov/postParser/internal/adapters/driver/tps"
	"github.com/vynovikov/postParser/internal/logger"
)

var (
	wgMain sync.WaitGroup
)

func main() {
	t := rpc.NewTransmitter(nil)
	s := store.NewStore()

	app, done := application.NewAppFull(s, t)

	tpR := tp.NewTpReceiver(app)
	tpsR := tps.NewTpsReceiver(app)

	go SignalListen(tpR, tpsR, app)
	go app.Start()
	go tpR.Run()
	go tpsR.Run()

	<-done
	logger.L.Errorln("postParser is interrupted")
}

// SignalListen listens for Interrupt signal, when receiving one invokes stop function
func SignalListen(tpR tp.TpReceiver, tpsR tps.TpsReceiver, app application.Application) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	<-sigChan
	Stop(tpR, tpsR, app)

}

// Stop sets stopping flog and invokes stop goroutines
func Stop(tpR tp.TpReceiver, tpsR tps.TpsReceiver, app application.Application) {
	app.SetStopping()
	wgMain.Add(2)
	go tpR.Stop(&wgMain)
	go tpsR.Stop(&wgMain)
	wgMain.Wait()
	app.Stop() // closes done in the end
}
