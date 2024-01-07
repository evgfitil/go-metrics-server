.PHONY: all server agent
all: server agent
server:
	cd cmd/server && go build -o server *.go

agent:
	cd cmd/agent && go build -o agent *.go