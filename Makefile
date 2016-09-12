TEAMID ?= 7357
repo=tompscanlan/q3reservation
bin=q3reservation

all: docker
local: $(bin)-local

$(bin): deps
	env GOOS=linux GOARCH=amd64 go build -a -v --installsuffix cgo  ./cmd/$(bin)

$(bin)-local: deps
	go build -a -v --installsuffix cgo  -o $(bin)-local  ./cmd/$(bin)

deps:
	go get -v ./...

docker: $(bin)
	docker build -t $(repo) --rm=true .

dockerclean: stop
	echo "Cleaning up Docker Engine before building."
	docker rm $$(docker ps -a | awk '/$(bin)/ { print $$1}') || echo -
	docker rmi $(repo)

clean: stop dockerclean
	go clean
	rm -f $(bin)

run:
	docker run -d -p8082:8082 -e BLOB_ID=$(TEAMID)  $(repo)

stop:
	docker kill $$(docker ps -a | awk '/$(bin)/ { print $$1}') || echo -

valid:
	go tool vet .
	go test -v -race ./...


.PHONY: imports docker clean run stop deps

