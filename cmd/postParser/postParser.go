package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"workspaces/postParser/internal/adapters/application"
	"workspaces/postParser/internal/adapters/driven/rpc"
	"workspaces/postParser/internal/adapters/driven/store"
	"workspaces/postParser/internal/adapters/driver/tp"
	"workspaces/postParser/internal/adapters/driver/tps"
	"workspaces/postParser/internal/core"
	"workspaces/postParser/internal/logger"
)

var (
	wgMain sync.WaitGroup
)

func main() {

	t := rpc.NewTransmitter(nil)
	c := core.NewCore()
	s := store.NewStore()

	app, done := application.NewAppFull(c, s, t)

	//logger.L.Infof("main.main app was set to %v\n", app)

	tpR := tp.NewTpReceiver(app)
	tpsR := tps.NewTpsReceiver(app)

	go signalListen(tpR, tpsR, app)
	go tpR.Run()
	go tpsR.Run()

	<-done
	logger.L.Errorln("in main.main done is closed, finishing")
}

func signalListen(tpR tp.TpReceiver, tpsR tps.TpsReceiver, app application.Application) {
	//logger.L.Infoln("main.signalListen invoked")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	//logger.L.Infof("in main.signalListen, <-sigChan: %v\n", <-sigChan)
	<-sigChan
	stop(tpR, tpsR, app)

}

func stop(tpR tp.TpReceiver, tpsR tps.TpsReceiver, app application.Application) {
	//logger.L.Infof("main.stop invoked with app = %v\n", app)
	app.SetStopping()
	wgMain.Add(2)
	go tpR.Stop(&wgMain)
	go tpsR.Stop(&wgMain)
	wgMain.Wait()
	//logger.L.Infof("in main.stop stopping app %v\n", app)
	app.Stop() // done is closed there

}
