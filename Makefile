BINARY := cmsc
BIN_DIR := ./sample
CMD := ./cmd/cms-cli

DIST_TARGETS := \
	linux/amd64/tar.gz \
	linux/arm64/tar.gz \
	darwin/amd64/tar.gz \
	darwin/arm64/tar.gz \
	windows/amd64/exe

.PHONY: fmt run build dist

fmt:
	go fmt ./...

run:
	go run $(CMD) $(RUN_ARGS)

build:
	mkdir -p $(BIN_DIR)
	rm -r $(BIN_DIR)/$(BINARY) || true
	go build -o $(BIN_DIR)/$(BINARY) $(CMD)

dist:
	@for target in $(DIST_TARGETS); do \
		GOOS=$$(echo $$target | cut -d/ -f1) \
		GOARCH=$$(echo $$target | cut -d/ -f2) \
		ARCHIVE=$$(echo $$target | cut -d/ -f3) \
		bash scripts/build.sh; \
	done
