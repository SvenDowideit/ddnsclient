
cloudflare: build
	docker run --rm -it \
		-v $(CURDIR)/ddnsclient.ini:/etc/ddnsclient.ini \
		ddnsclient -config=/etc/ddnsclient.ini -ip 33.44.55.22

help: build
	docker run --rm -it ddnsclient

build:
	docker build -t ddnsclient .
