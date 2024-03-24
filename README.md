# Go station 

### Motivation: learn Golang

Go station is a server built to take advantage of Go's ease-of-use concurrency execution, through the use of goroutines and channels. 

To start up a server:

```
go run ./server <server_port> <file 1> <file 2> ... 
```


To start up a client controller:

```
go run ./client/control <ip-addr> <server_port> <udp-port>
```

To start up a client listener
```
go run ./client/listener <udp-port>
```

alternatively, you can run this command to output the audio from your device's speakers

```
go run ./client/listener <udp-port> | mpg123 -  
```
