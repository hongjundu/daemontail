package main

import (
	"flag"
	"github.com/hongjundu/go-level-logger"
	"github.com/hpcloud/tail"
	"github.com/sevlyar/go-daemon"
	"log"
	"os"
	"syscall"
	"time"
)

var (
	signal = flag.String("s", "", `Send signal to the daemon:
  quit — graceful shutdown
  stop — fast shutdown
  reload — reloading the configuration file`)

	stop = make(chan struct{})
	done = make(chan struct{})
)

func main() {
	flag.Parse()
	daemon.AddCommand(daemon.StringFlag(signal, "quit"), syscall.SIGQUIT, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)
	daemon.AddCommand(daemon.StringFlag(signal, "reload"), syscall.SIGHUP, reloadHandler)

	logger.Init(0, "mytail", "/tmp/", 100, 3, 30)
	logger.Debugf("[main] enters args: %v", os.Args)

	cntxt := &daemon.Context{
		PidFileName: "/tmp/mytail.pid",
		PidFilePerm: 0644,
		LogFileName: "/tmp/mytail2.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        nil,
	}

	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			logger.Fatalf("Unable send signal to the daemon: %s", err.Error())
		}
		daemon.SendCommands(d)
		return
	}

	d, err := cntxt.Reborn()
	if err != nil {
		logger.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	go func() {
		t, err := tail.TailFile("/tmp/mytail.log", tail.Config{Follow: true})

		if err != nil {
			logger.Errorln(err)
			return
		}

		for line := range t.Lines {
			log.Println("tail reading: >>> " + line.Text)
		}
	}()

	go func() {
		i := 1
	LOOP:
		for {
			select {
			case <-stop:
				break LOOP
			default:
			}

			logger.Infof("log line : %d", i)
			time.Sleep(time.Second)
			i++
		}

		done <- struct{}{}
	}()

	err = daemon.ServeSignals()
	if err != nil {
		logger.Errorf("Error: %s", err.Error())
	}

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	//quit := make(chan os.Signal)

	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	//signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	//<-quit

	logger.Infof("Server exited")

}

func termHandler(sig os.Signal) error {
	logger.Infoln("terminating...")
	stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}

func reloadHandler(sig os.Signal) error {
	logger.Infoln("configuration reloaded")
	return nil
}
