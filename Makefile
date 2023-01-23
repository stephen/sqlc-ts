internal/plugin/codegen.pb.go: protos/plugin/codegen.proto
	protoc -I ./protos \
		--go_out=. \
		--go_opt=module=github.com/stephen/sqlc-ts \
		--go-vtproto_out=. \
		--go-vtproto_opt=module=github.com/stephen/sqlc-ts,features=marshal+unmarshal+size \
		./protos/plugin/codegen.proto
