internal/plugin/codegen.pb.go: protos/plugin/codegen.proto
	protoc -I ./protos \
		--go_out=. \
		--go_opt=module=github.com/stephen/sqlc-sql.js \
		--go-vtproto_out=. \
		--go-vtproto_opt=module=github.com/stephen/sqlc-sql.js,features=marshal+unmarshal+size \
		./protos/plugin/codegen.proto
