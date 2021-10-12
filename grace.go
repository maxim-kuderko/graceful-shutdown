package graceful_shutdown

import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/atomic"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const gracefulShutdownTimeoutENV = `GRACEFUL_SHUTDOWN_TIMEOUT`
const healthPORT = `HEALTH_PORT`
const gracefulShutdownTimeoutDefault = time.Second * 1

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
	wgIsDown.Add(1)
	wgShuttingDown.Add(1)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go waitForSignal(c)
	if os.Getenv(healthPORT) != `` {
		go http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv(healthPORT)), healthHandler{}) //nolint
	}
}

func waitForSignal(c chan os.Signal) {
	duration := gracefulShutdownTimeoutDefault
	v := viper.New()
	v.AutomaticEnv()
	if tmp := v.GetDuration(gracefulShutdownTimeoutENV) * time.Second; tmp > time.Second {
		duration = tmp
	}
	<-c
	wgShuttingDown.Done()
	alive.Store(false)
	time.Sleep(duration)
	wgIsDown.Done()
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
