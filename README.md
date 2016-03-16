# dDNSclient

Set the IP address of a host.

Currently supports:

* [`cloudflare`](https://www.cloudflare.com), create or update a host record
  Use the Clourflare API token as `-password`
* [`noip`](http://www.noip.com), update only
* [`dreamhost`](http://dreamhost.com),
  Use an API toke generated using https://panel.dreamhost.com/?tree=home.api
  Note that the Dreamhost DNS is not really good for ddns - update requires a remove and add operation, with weird side effects.
  Its still useful for adding a new `A` record :)

```
ddnsclient  -debug -verbose \
    -protocol noip -host something.ddns.net \
    -login=someone@gmail.com -password=something
    -ip 66.66.66.88
```

OR you can set the values in a config file:

```
$ ddnsclient -config ddnsclient.ini -ip=22.33.44.55
Set baloo.demo.gallery to 22.33.44.55
Currently set baloo.demo.gallery to 66.66.66.77
Set baloo.demo.gallery to 22.33.44.55
```

using `ddnsclient.ini`:

```
# docker run --rm -it -v $(pwd)/ddnsclient.ini:/etc/ddnsclient.ini ddnsclient -config=/etc/ddnsclient.i
ni
host=baloo.demo.gallery
protocol=cloudflare
login=svendowideit@demo.gallery
password=<API token>
ip=66.66.66.77
```