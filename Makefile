start-client:
	@go run client/client.go

start-server:
	@go run server/server.go

clean-db:
	@sudo rm -rf database.db
