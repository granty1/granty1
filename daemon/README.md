# daemon


## example
```
» go get -u github.com/granty1/granty1
```

```
import "github.com/granty1/granty1/daemon"

func main() {
    // start daemon process
    // before other flag parse!
    daemon.Run()
}
```
```
go build -o server .

`pwd`/server -d
```
```
» daemon proc start success!
```