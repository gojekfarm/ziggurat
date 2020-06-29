### Ziggurat GO [WIP]

#### A Golang prototype of Ziggurat

- The Ziggurat dir contains the files which will be bundled as a go pkg.
- The driver file `main.go` shows how the public API might look like.

#### Requirements to run the spike
    Go 1.14 https://golang.org/doc/install
    Docker https://docs.docker.com/get-docker/

#### How to run the spike

- Run the command `make run-spike` in a different terminal OR tab
-   Run the driver file `main.go`
```shell script
go run main.go
```
- You should be able to see the handler func logging for every message received