GO           ?= go
LINTER       ?= golangci-lint
ALIGNER      ?= betteralign
VULNCHECK    ?= govulncheck
MARKDOWNLINT ?= rumdl
BENCHSTAT    ?= benchstat
BENCH_COUNT  ?= 6
BENCH_REF    ?= bench_baseline.txt

DOC_SOURCE_DATE_EPOCH ?= 1700000000

TARGETS ?= \
	linux/arm/7 \
	linux/arm/5 \
	windows/386 \
	windows/amd64 \
	darwin/amd64 \
	freebsd/amd64 \
	aix/ppc64 \
	solaris/amd64

.PHONY: test test-race test-short bench bench-fast bench-reset verify vet check ci \
	fmt fmt-check lint lint-fix align align-fix tidy tidy-check download vulncheck markdownlint \
	tools tools-ci tool-golangci-lint tool-betteralign tool-govulncheck tool-benchstat \
	release-notes crosscompile docs-render docs-render-check

check: verify vulncheck tidy fmt vet lint-fix align-fix test docs-render markdownlint
ci: download tools-ci verify vulncheck tidy-check fmt-check vet lint align test docs-render-check

fmt:
	gofmt -w .

fmt-check:
	@files=$$(gofmt -l .); \
	if [ -n "$$files" ]; then \
		echo "$$files" 1>&2; \
		echo "gofmt: files need formatting" 1>&2; \
		exit 1; \
	fi

vet:
	$(GO) vet ./...

test:
	$(GO) test .

test-race:
	$(GO) test -race .

test-short:
	$(GO) test -short .

bench:
	@tmp=$$(mktemp); \
	$(GO) test ./... -run=^$$ -bench 'Benchmark' -benchmem -count=$(BENCH_COUNT) | tee "$$tmp"; \
	if [ -f "$(BENCH_REF)" ]; then \
		$(BENCHSTAT) "$(BENCH_REF)" "$$tmp"; \
	else \
		cp "$$tmp" "$(BENCH_REF)" && echo "Baseline saved to $(BENCH_REF)"; \
	fi; \
	rm -f "$$tmp"

bench-fast:
	$(GO) test ./... -run=^$$ -bench 'Benchmark' -benchmem

bench-reset:
	rm -f "$(BENCH_REF)"

verify:
	$(GO) mod verify

tidy-check:
	@$(GO) mod tidy
	@git diff --stat --exit-code -- go.mod go.sum || ( \
		echo "go mod tidy: repository is not tidy"; \
		exit 1; \
	)

tidy:
	$(GO) mod tidy

download:
	$(GO) mod download

lint:
	$(LINTER) run ./...

lint-fix:
	$(LINTER) run --fix ./...

align:
	$(ALIGNER) ./...

align-fix:
	-$(ALIGNER) -apply ./...
	$(ALIGNER) ./...

vulncheck:
	$(VULNCHECK) ./...

markdownlint:
	@if command -v $(MARKDOWNLINT) >/dev/null 2>&1; then \
		$(MARKDOWNLINT) check; \
	else \
		echo "WARN: $(MARKDOWNLINT) not found; skipping markdown lint."; \
		echo 'WARN: Install it https://github.com/rvben/rumdl'; \
	fi

tools: tool-golangci-lint tool-betteralign tool-govulncheck tool-benchstat
tools-ci: tool-golangci-lint tool-betteralign tool-govulncheck

tool-golangci-lint:
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

tool-betteralign:
	$(GO) install github.com/dkorunic/betteralign/cmd/betteralign@latest

tool-govulncheck:
	$(GO) install golang.org/x/vuln/cmd/govulncheck@latest

tool-benchstat:
	$(GO) install golang.org/x/perf/cmd/benchstat@latest

release-notes:
	@awk '\
	/^<!--/,/^-->/ { next } \
	/^## \[[0-9]+\.[0-9]+\.[0-9]+\]/ { if (found) exit; found=1; next } \
	found { \
		if (/^## \[/) { exit } \
		if (/^$$/) { flush(); print; next } \
		if (/^\* / || /^- /) { flush(); buf=$$0; next } \
		if (/^###/ || /^\[/) { flush(); print; next } \
		sub(/^[ \t]+/, ""); sub(/[ \t]+$$/, ""); \
		if (buf != "") { buf = buf " " $$0 } else { buf = $$0 } \
		next \
	} \
	function flush() { if (buf != "") { print buf; buf = "" } } \
	END { flush() } \
	' CHANGELOG.md

crosscompile:
	@set -e; \
	for target in $(TARGETS); do \
		IFS=/; set -- $$target; unset IFS; \
		goos=$$1; goarch=$$2; goarm=$$3; \
		if [ -n "$$goarm" ]; then \
			echo "# $$goos $$goarch v$$goarm"; \
			GOOS=$$goos GOARCH=$$goarch GOARM=$$goarm $(GO) build ./...; \
		else \
			echo "# $$goos $$goarch"; \
			GOOS=$$goos GOARCH=$$goarch $(GO) build ./...; \
		fi; \
	done

docs-render:
	SOURCE_DATE_EPOCH=$(DOC_SOURCE_DATE_EPOCH) UPDATE_DOC_EXAMPLES=1 \
	$(GO) test -tags forceposix -run TestWriteDocBuiltinTemplatesGolden ./...

docs-render-check:
	SOURCE_DATE_EPOCH=$(DOC_SOURCE_DATE_EPOCH) UPDATE_DOC_EXAMPLES=1 \
	$(GO) test -tags forceposix -run TestWriteDocBuiltinTemplatesGolden ./...
	git diff --exit-code -- examples/doc-rendered
