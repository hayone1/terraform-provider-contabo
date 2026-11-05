ifndef VERSION
	VERSION = 'v0.0.1'
endif

# The Go API client in ./openapi is generated from a committed snapshot of the
# Contabo OpenAPI spec (spec/api-v1.yaml), so builds are reproducible and do not
# depend on the live spec. Run `make update-spec` to refresh the snapshot, then
# `make generate-api-clients` and commit both.
OPENAPIURL ?= https://api.contabo.com/api-v1.yaml
SPEC ?= spec/api-v1.yaml
NORMALIZED_SPEC = build/openapi-spec.yaml
OPENAPI_GENERATOR_IMAGE ?= openapitools/openapi-generator-cli:v5.2.1
YQ_IMAGE ?= mikefarah/yq:4.44.3
DOCKER_USER ?= $(shell id -u):$(shell id -g)

.PHONY: test
test:
	go test ./...

.PHONY: test-acc
test-acc:
	TF_ACC=1 go test -v ./...

.PHONY: build
build: build-only

.PHONY: update-spec
update-spec:
	curl -sSfL $(OPENAPIURL) -o $(SPEC)

# Operations with multiple tags are generated into each tag's API file, which
# causes duplicate declarations in the Go client. Keep only the first tag.
.PHONY: generate-api-clients
generate-api-clients:
	rm -rf openapi build
	mkdir -p build
	docker run --rm -i $(YQ_IMAGE) \
		'(.paths[][] | select(tag == "!!map" and has("tags")) | .tags) |= .[0:1]' - < $(SPEC) > $(NORMALIZED_SPEC)
	docker run --rm --user $(DOCKER_USER) -v "$(CURDIR):/local" $(OPENAPI_GENERATOR_IMAGE) generate \
		--skip-validate-spec \
		--input-spec /local/$(NORMALIZED_SPEC) \
		--generator-name go \
		--package-name openapi \
		--global-property apiDocs=false,modelDocs=false,apiTests=false,modelTests=false \
		--output /local/openapi
	rm -rf openapi/api openapi/.travis.yml openapi/git_push.sh openapi/.openapi-generator-ignore
	cd openapi && go mod edit -module contabo.com/openapi && gofmt -w .
	rm -rf build

# Fails when the committed client differs from what the spec snapshot generates.
.PHONY: check-api-clients
check-api-clients: generate-api-clients
	git diff --exit-code -- openapi
	@test -z "$$(git status --porcelain -- openapi)" || (git status --porcelain -- openapi; exit 1)

.PHONY: docs
docs:
	go generate ./...

.PHONY: check-docs
check-docs: docs
	git diff --exit-code -- docs
	@test -z "$$(git status --porcelain -- docs)" || (git status --porcelain -- docs; exit 1)

.PHONY: lint
lint:
	go vet ./...
	golangci-lint run

.PHONY: build-only
build-only:
	go build -o terraform-provider-contabo_$(VERSION)

.PHONY: build-only-debug
build-only-debug:
	go build -gcflags="all=-N -l" -o terraform-provider-contabo_$(VERSION)

.PHONY: doc-preview
doc-preview:
	@echo "Preview your markdown documentation on this page: https://registry.terraform.io/tools/doc-preview"
