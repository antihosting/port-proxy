# port-proxy

Simple Port Proxy. Routes traffic from one port to another. Primary usage is to run proxy on lower ports and forward traffic for high ports.

### Build

Before compiling you should create file `.hashed-token.txt` and place this token there:
```
eM4OLAhhhZ6miCXUCMRmtSc9QbxfPi9CQ0tduGsMkhY
```

That corresponds to this set:
```
Token: dTr9uDeunWizqUKf4dCsvklj46ylKQnSiTntT7ryhHs
Hashed Token: eM4OLAhhhZ6miCXUCMRmtSc9QbxfPi9CQ0tduGsMkhY
```

You can also generate your own tokens by running this command:
```
./port_proxy -g
```

If you do not want to enter token every time you run proxy, create env:
```
export PORT_PROXY_TOKEN=dTr9uDeunWizqUKf4dCsvklj46ylKQnSiTntT7ryhHs
```

Tokens are using to protect your server BIND functionality, especially if you setup low port permissions to this proxy.
Which is the common ussage for this proxy.
```
sudo setcap CAP_NET_BIND_SERVICE=+eip port_proxy_linux
```

In order to build executable for your OS, run this command:
```
make
```

In order to build ditributives for `linux,mac,windows`, run this command:
```
make distr
```

### Usage

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

Socket benchmark:
```
./port_proxy -b socket -ip 127.0.0.1 -p 40551:40561
```

HTTP benchmark:
```
./port_proxy -b http -ip 127.0.0.1 -p 40551:40561
```


