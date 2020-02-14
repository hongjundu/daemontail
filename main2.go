package main

import (
	"flag"
	"github.com/hpcloud/tail"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	file := flag.String("f", "", "Text file")
	flag.Parse()

	if len(*file) == 0 {
		log.Println("no file specifed")
		return
	}

	go func() {
		t, err := tail.TailFile(*file, tail.Config{Follow: true, Poll: true})

		if err != nil {
			log.Println(err)
			return
		}

		for line := range t.Lines {
			log.Println("tail reading: >>> " + line.Text)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)

	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server exited")

}
