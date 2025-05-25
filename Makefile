# base path for Lexicon document tree (for lexgen)
LEXDIR?=./pkg/atproto/vtri

# Proto related variables
PROTO_DIR=./proto/chat
GO_OUT_DIR=./proto/chat
PROTO_FILES=$(shell find $(PROTO_DIR) -name "*.proto")

.PHONY: install-proto-tools
install-proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: proto
proto: install-proto-tools
	@echo "Generating proto files..."
	@for file in $(PROTO_FILES); do \
		protoc \
			--proto_path=$(PROTO_DIR) \
			--go_out=$(GO_OUT_DIR) \
			--go_opt=paths=source_relative \
			--go-grpc_out=$(GO_OUT_DIR) \
			--go-grpc_opt=paths=source_relative \
			$$file; \
	done
	@echo "Proto files generated successfully"

# 清理生成的proto文件
.PHONY: clean-proto
clean-proto:
	find $(GO_OUT_DIR) -name "*.pb.go" -type f -delete

.PHONY: lexgen
lexgen:
	go run github.com/bluesky-social/indigo/cmd/lexgen --package vtri \
	    --types-import app.vtri:github.com/zhongshangwu/avatarai-social \
	    --outdir ./pkg/atproto/vtri \
	    --prefix app.vtri \
	    --build-file ./lexgen-build.json \
	    pkg/atproto/lexicons \
	    ../atproto/lexicons

.PHONY: go-lexicons
go-lexicons:
	rm -rf ./pkg/atproto/vtri \
	&& mkdir -p ./pkg/atproto/vtri \
	&& rm -rf ./pkg/atproto/vtri/cbor_gen.go \
	&& $(MAKE) lexgen \
	&& sed -i.bak 's/\tutil/\/\/\tutil/' $$(find ./pkg/atproto/vtri -type f) \
	&& sed -i.bak '/func .*MarshalCBOR\|func .*UnmarshalCBOR/,/^}/ s/^/\/\//' $$(find ./pkg/atproto/vtri -type f) \
	&& go run golang.org/x/tools/cmd/goimports@latest -w $$(find ./pkg/atproto/vtri -type f) \
	&& go run ./cmd/atpgen/main.go \
	&& $(MAKE) lexgen \
	&& rm -rf ./pkg/atproto/vtri/*.bak \
	&& rm -rf api
	&& sed -i '/func .*MarshalCBOR\|func .*UnmarshalCBOR/,/^}/ s/^/\/\//' $$(find ./pkg/atproto/vtri -type f|grep -v cbor_gen.go) \
	&& go run golang.org/x/tools/cmd/goimports@latest -w $$(find ./pkg/atproto/vtri -type f|grep cbor_gen.go)


.PHONY: go-lexicons
go-lexicons:
	rm -rf ./pkg/atproto/vtri \
	&& mkdir -p ./pkg/atproto/vtri \
	&& rm -rf ./pkg/atproto/vtri/cbor_gen.go \
	&& $(MAKE) lexgen \
	&& find ./pkg/atproto/vtri -type f | xargs sed -i.bak 's/\tutil/\/\/\tutil/' \
	&& find ./pkg/atproto/vtri -type f | xargs sed -i.bak '/func .*MarshalCBOR\|func .*UnmarshalCBOR/,/^}/ s/^/\/\//' \
	&& go run golang.org/x/tools/cmd/goimports@latest -w ./pkg/atproto/vtri \
	&& go run ./cmd/atpgen/main.go \
	&& $(MAKE) lexgen \
	&& find ./pkg/atproto/vtri -type f | grep -v cbor_gen.go | xargs sed -i '/func .*MarshalCBOR\|func .*UnmarshalCBOR/,/^}/ s/^/\/\//' \
	&& go run golang.org/x/tools/cmd/goimports@latest -w ./pkg/atproto/vtri \
	&& rm -rf ./pkg/atproto/vtri/*.bak \
	&& rm -rf api


# .PHONY: lexgen
# lexgen: ## Run codegen tool for lexicons (lexicon JSON to Go packages)
# 	go run ./cmd/lexgen/ --build-file cmd/lexgen/lexgen-build.json $(LEXDIR)

# .PHONY: cborgen
# cborgen: ## Run codegen tool for CBOR serialization
# 	go run ./gen
