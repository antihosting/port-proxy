# tcp-proxy

Simple TCP Proxy. Routes traffic from one port to another. Primary usage is to run proxy on lower ports and forward traffic for higher ports.

### Build

In order to build executable for your OS, run this command:
```
make
```

In order to build executables for `linux,mac,windows` run this command:
```
make target
```

### Usage

This is the common usage for this proxy. Allow only this process for `setcap` and redirect traffic to higher ports.
```
sudo setcap CAP_NET_BIND_SERVICE=+eip tcp_proxy_linux
```

Start proxy in background:
```
./port_proxy -ip 127.0.0.1 -p 40551:40561 -p 40552:40562
```

Start proxy in foreground:
```
./port_proxy -f -ip 127.0.0.1 -p 40551:40561 -p 40552:40562
```

Verbose logs:
```
./port_proxy -f -v -ip 127.0.0.1 -p 40551:40561
```

### Benchmarks

Proxy supports HTTP and socket benchmarks embedded in it, so you can test on server performance before deployment.

Socket benchmark on already running proxy from 40551 to 40561:
```
./port_proxy -b socket -ip 127.0.0.1 -p 40551:40561
```

Socket benchmark on our proxy:
```
./port_proxy -b socket-proxy -ip 127.0.0.1 -p 40551:40561
```

HTTP benchmark on already running proxy from 40551 to 40561:
```
./port_proxy -b http -ip 127.0.0.1 -p 40551:40561
```

HTTP benchmark on our proxy:
```
./port_proxy -b http-proxy -ip 127.0.0.1 -p 40551:40561
```

### Benchmarks Results

I was not able to find faster proxy on the market than this one.

### Problems

Dial could be slow on Linux systems after restart of backend.
Nothing related to this proxy, looks like OS system issue or golang itself. 

