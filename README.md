Simple implementation

1. set envs 
   ```env 
   HEALTH_PORT=<PORT> // i.e 9999
   GRACEFUL_SHUTDOWN_TIMEOUT=30 // in seconds
   ```
   
2. in code

```go
import  gs "github.com/maxim-kuderko/graceful-shutdown"

func main(){
	go func{
		...application code
    }()
    gs.WaitForGrace()	
}
```

optional: use the gs.ShuttingDownHook() to start closing any background operation

```go
import  gs "github.com/maxim-kuderko/graceful-shutdown"

func main(){
	c := make(chan struct{}, 1)
	go processC(c)
	
	gs.ShuttingDownHook()
	close(c)
    gs.WaitForGrace()	
}

func processC(c <-chan struct{}){
	..background operation...
}
```