.PHONY: build clean stop start restart test

build:
	go build -o mini-kanban ./cmd

clean:
	rm -f mini-kanban srv/srv

test:
	go test ./...
