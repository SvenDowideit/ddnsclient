


BUILD_DATE := $(shell date +%Y-%m-%d\-%H-%M)
BRANCH := $(shell git branch | grep "^*" | sed 's/* //g')
REV_INFO := $(shell git describe --all --dirty --tags | sed "s/\//-/g")
VERSION :=$(REV_INFO)-$(BUILD_DATE)

cloudflare: build
	docker run --rm -it \
		-v $(CURDIR)/ddnsclient.ini:/etc/ddnsclient.ini \
		ddnsclient -config=/etc/ddnsclient.ini -ip 33.44.55.22

help: build
	docker run --rm -it ddnsclient

build:
	docker build -t ddnsclient --build-arg VERSION=$(VERSION) .

release: create-release-tag checkout-tag build create-github-release

DIRTY:=$(shell git diff --shortstat 2> /dev/null | tail -n1)
REPO:=SvenDowideit/ddnsclient
MASTER_BRANCH_SHA:=$(shell curl https://$(GITHUB_TOKEN)@api.github.com/repos/$(REPO)/git/refs/heads/master  2>/dev/null | grep sha | sed 's/.*"sha": "\(.*\)".*/\1/')

create-release-tag:
ifneq ($(DIRTY),)
	$(error Please make sure local checkout is clean - .)
endif
	# TODO: should really check to see if the MASTER_BRANCH_SHA has already been tagged
	echo "Creating the $(VERSION) tag from master $(MASTER_BRANCH_SHA)"
	curl -X POST --data \
		'{"ref":"refs/tags/$(VERSION)", "sha":"$(MASTER_BRANCH_SHA)"}' \
		https://$(GITHUB_TOKEN)@api.github.com/repos/$(REPO)/git/refs

checkout-tag:
	git fetch --all
	git checkout -b $(VERSION)

create-github-release:
	echo "Creating the $(VERSION) release using tag"
	curl -X POST --data \
		'{"tag_name": "$(VERSION)","name": "$(VERSION)","body": "we built a release","draft": true,"prerelease": true}' \
		https://$(GITHUB_TOKEN)@api.github.com/repos/$(REPO)/releases
	echo "Now upload the ddclient binary"
	# TODO: uploading is more fun.
