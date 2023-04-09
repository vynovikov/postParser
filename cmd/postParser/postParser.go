package main

import (
	"os"
	"os/signal"
	"postParser/internal/adapters/application"
	"postParser/internal/adapters/driven/rpc"
	"postParser/internal/adapters/driven/store"
	"postParser/internal/adapters/driver/tp"
	"postParser/internal/adapters/driver/tps"
	"postParser/internal/core"
	"postParser/internal/logger"
	"sync"
	"syscall"
)

var (
	wgMain sync.WaitGroup
)

func main() {
	//logger.L.Infoln("in main.main NewTransmitter")
	t := rpc.NewTransmitter(nil)
	//logger.L.Infoln("in main.main NewCore")
	c := core.NewCore()
	//logger.L.Infoln("in main.main NewStore")
	s := store.NewStore()

	//logger.L.Infoln("in main.main NewAppFull")
	app, done := application.NewAppFull(c, s, t)

	//logger.L.Infof("main.main app was set to %v\n", app)

	//logger.L.Infoln("in main.main NewTpReceiver")
	tpR := tp.NewTpReceiver(app)
	//logger.L.Infoln("in main.main NewTpsReceiver")
	tpsR := tps.NewTpsReceiver(app)

	go signalListen(tpR, tpsR, app)
	go app.Do()
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
