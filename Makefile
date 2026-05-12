.DEFAULT_GOAL := help

.PHONY: help build clean distclean test fmt vet check golden-version golden-json golden-tree

help:
	@echo "Targets:"
	@echo "  build            go build -> ./bin/exiftool-go"
	@echo "  test             go test ./... (set EXIFTOOL_GO_EXIFTOOL or use exiftool on PATH)"
	@echo "  fmt              go fmt ./..."
	@echo "  vet              go vet ./..."
	@echo "  check            fmt, vet, test"
	@echo "  clean            remove ./bin"
	@echo "  distclean        same as clean"
	@echo "  golden-version   regen testdata/golden/version.stdout"
	@echo "  golden-json      regen testdata/golden/json_samsung_sm_g930p.json"
	@echo "  golden-tree      regen testdata/golden/tree_r_filename_sorted.stdout"

build:
	mkdir -p bin
	go build -o bin/exiftool-go ./cmd/exiftool-go

test:
	go test ./... -count=1 -timeout=15m -cover

fmt:
	go fmt ./...

vet:
	go vet ./...

check: fmt vet test

clean:
	rm -rf bin

distclean: clean

golden-version:
	./scripts/exiftoolgen version

golden-json:
	./scripts/exiftoolgen json-samsung

golden-tree:
	./scripts/exiftoolgen tree-recurse
