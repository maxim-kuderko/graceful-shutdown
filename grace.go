package graceful_shutdown

import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/atomic"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

const gracefullShutdownTimeoutENV = `GRACEFUL_SHUTDOWN_TIMEOUT`
const healthPORT = `HEALTH_PORT`
const gracefullShutdownTimeoutDefault = time.Second * 1

var (
	wgIsDown       = sync.WaitGroup{}
	wgShuttingDown = sync.WaitGroup{}

	WaitForGrace = func() {
		wgIsDown.Wait()
	}

	ShuttingDownHook = func() {
		wgShuttingDown.Wait()
	}
	alive   = atomic.NewBool(true)
	IsAlive = func() bool {
		return alive.Load()
	}
)

func init() {
	if os.Getenv(healthPORT) == `` {
		log.Println(healthPORT, ` is not set`)
	}
	wgIsDown.Add(1)
	wgShuttingDown.Add(1)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		duration := gracefullShutdownTimeoutDefault
		if tmp := viper.New().GetDuration(gracefullShutdownTimeoutENV); tmp > time.Second {
			duration = tmp
		}
		<-c
		wgShuttingDown.Done()
		alive.Store(false)
		time.Sleep(duration)
		wgIsDown.Done()
	}()

	go func() {
		http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv(healthPORT)), healthHandler{}) //nolint
	}()
}

type healthHandler struct {
}

func (h healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !IsAlive() {
		w.WriteHeader(http.StatusLocked)
		return
	}
	w.Write([]byte(`OK`))
}
