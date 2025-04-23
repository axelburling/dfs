.PHONY: generate
generate:
	sqlc generate

.PHONY: node-proto
node-proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./pkg/node/grpc/pb/node.proto

.PHONY: master-proto
master-proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./pkg/master/grpc/pb/master.proto

.PHONY: pubsub-proto
pubsub-proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./pkg/pubsub/grpc/pb/pubsub.proto

.PHONY: migrate-up
migrate-up:
	sh migrate.sh up

.PHONY: migrate-down
migrate-down:
	sh migrate.sh down
