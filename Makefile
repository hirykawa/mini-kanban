.PHONY: build clean stop start restart test build-web

build-web:
	cd web && npm install && npm run build

build: build-web
	go build -o mini-kanban ./cmd

clean:
	rm -f mini-kanban srv/srv

test:
	go test ./...
