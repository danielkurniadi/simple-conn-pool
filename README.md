# Connection Pool in Go

This repository contains example of managing a connection pool in Golang. Currently, these are the variants and you can use depending
on your use cases:

1. Queue Connection Pool (simple)
2. ...


## Install and Usage

Install the package with:

```bash
go get github.com/fatih/pool
```

Please vendor the package with one of the releases: https://github.com/fatih/pool/releases.
`master` branch is **development** branch and will contain always the latest changes.


## Example

### 1. Queue Connection Pool (simple)

```go
// create a factory() to be used with channel based pool
constructor := func() (net.Conn, error) { return net.Dial("tcp", "127.0.0.1:3360") }

// create a new channel based pool with an initial capacity of 5 and maximum
// capacity of 30. The factory will create 5 initial connections and put it
// into the pool.
queuePool, err := pool.NewQueuePool(5, 30, constructor)

// now you can get a connection from the pool, if there is no connection
// available it will create a new one via the factory function.
conn, err := queuePool.Get()

// do something with conn
doSomething(conn, ...)

// when you are done, closing the connection
// will put it back in the queue instead, but not
// closing it so that we can reuse this conn

// close pool any time you want, this closes all the connections inside a pool
queuePool.Close()

// currently available connections in the pool
currentNum := queuePool.Len()
```