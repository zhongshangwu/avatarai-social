# base path for Lexicon document tree (for lexgen)
LEXDIR?=./pkg/atproto/vtri

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

# .PHONY: lexgen
# lexgen: ## Run codegen tool for lexicons (lexicon JSON to Go packages)
# 	go run ./cmd/lexgen/ --build-file cmd/lexgen/lexgen-build.json $(LEXDIR)

# .PHONY: cborgen
# cborgen: ## Run codegen tool for CBOR serialization
# 	go run ./gen
