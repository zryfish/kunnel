## Kunnel
Kunnel is short for **Kubernetes tunnel**, built for exposing Kubernetes service to outside the cluster. It requires a globally accessible domain and creates a reverse tunnel from that domain to Kubernetes.

## Install

### Binaries

### Build from source
```
git clone https://github.com/zryfish/kunnel.git
cd kunnel
make all
```
Binaries `server` and `client` will be found under directory `bin/`.

## How to run
### Run Locally
A top-level domain is required to run the server. For testing purposes, you can specify any domain.
```shell
./server  --domain kunnel.run --port 80
# I0901 17:49:31.522742   45829 main.go:37] server started
```

Now, the server is ready to accept requests. Suppose there is a service running on `http://192.168.0.12:8000`, run client as follows, rem:
```shell
./client --server ws://localhost:80 --local 192.168.0.12:8000
# I0901 17:54:32.551148   48845 client.go:180] Service available at q0e9ioxrap.kunnel.run
```

We can test the tunnel using `curl`, 
```
$ curl -v --resolve q0e9ioxrap.kunnel.run:80:127.0.0.1 http://q0e9ioxrap.kunnel.run
* Added q0e9ioxrap.kunnel.run:80:127.0.0.1 to DNS cache
* Hostname q0e9ioxrap.kunnel.run was found in DNS cache
*   Trying 127.0.0.1...
* TCP_NODELAY set
* Connected to q0e9ioxrap.kunnel.run (127.0.0.1) port 80 (#0)
> GET / HTTP/1.1
> Host: q0e9ioxrap.kunnel.run
> User-Agent: curl/7.64.1
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Length: 13
< Content-Type: text/html
< Date: Wed, 01 Sep 2021 09:59:28 GMT
< Last-Modified: Wed, 01 Sep 2021 09:58:07 GMT
< Server: SimpleHTTP/0.6 Python/2.7.16
<
Hello World!
* Connection #0 to host q0e9ioxrap.kunnel.run left intact
* Closing connection 0
```

## Run on the server
A globally accessible domain is required to run server publicly, for example, kunnel.run. Also need to make sure the following DNS record exists on your domain provider.
```
A * [SERVER IP]
```
