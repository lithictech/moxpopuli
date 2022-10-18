ifdef DEBUG
	DEBUGARG := --log-level=debug
endif

XTRAARGS := ${DEBUGARG}
CMS := $(realpath ./cms)
BUILDFLAGS := -ldflags "-X github.com/lithictech/moxpopuli.BuildTime=`date -u +"%Y-%m-%dT%H:%M:%SZ"` -X github.com/lithictech/moxpopuli.BuildSha=`git rev-list -1 HEAD`"

guardcmd-%:
	@hash $(*) > /dev/null 2>&1 || \
		(echo "ERROR: '$(*)' must be installed and available on your PATH."; exit 1)

guardenv-%:
	@if [ -z '${${*}}' ]; then echo 'ERROR: environment variable $* not set' && exit 1; fi

fmt:
	@go fmt ./...

lint: guardcmd-gofmt
	@test -z $$(gofmt -d -l . | tee /dev/stderr) && echo "gofmt ok"

vet:
	@go vet

test:
	@LOG_LEVEL=fatal ginkgo -r

update-test-snapshots:
	UPDATE_SNAPSHOTS=true make test

bench:
	@LOG_LEVEL=fatal ginkgo -r --focus=benchmarks

build:
	@go build -o moxpopuli cmd/moxpopuli/main.go

install: build
	@mv ./moxpopuli `go env GOPATH`/bin

_mktemp:
	@mkdir -p .temp

t-whdb-localhost-schema: build _mktemp ## Read WebhookDB logged webhooks from localhost to generate a schema.
	./moxpopuli schemagen \
		--loader=file://./.temp/whdb-localhost.json \
		--saver=file://./.temp/whdb-localhost.json \
		--payload-loader=postgres://webhookdb:webhookdb@localhost:18005/webhookdb \
		--payload-loader-arg="SELECT request_body FROM logged_webhooks WHERE truncated_at IS NULL LIMIT 100" \

t-whdb-localhost-data: build _mktemp ## Generate data from the localhost logged webhooks schema.
	./moxpopuli datagen \
		--loader=file://./.temp/whdb-localhost.json

t-whdb-csv-schema: build _mktemp ## Read WebhookDB logged webhooks from an export CSV.
	./moxpopuli schemagen \
			--payload-loader=file://./../webhookdb-api/temp/logged-webhooks-stripecharges.json.csv \
    		--loader=file://./.temp/whdb-csv.json \
    		--saver=file://./.temp/whdb-csv.json \

t-whdb-csv-data: build _mktemp ## Use the schema generated from the CSV export to generate data.
	./moxpopuli datagen \
		--loader=file://./.temp/whdb-csv.json

t-fixtures-gen: build _mktemp ## Generate fixtures exercising all formats.
	./moxpopuli fixturegen --count=5 > .temp/fixtures.json

t-fixtures-schema: build _mktemp ## Generate schema from fixtures.
	./moxpopuli schemagen \
		--payload-loader=file://./.temp/fixtures.json \
		--saver=file://./.temp/fixtures.schema.json

t-whdb-spec-from-db: build _mktemp
	./moxpopuli specgen \
		--loader='file://./testdata/whdbspecseed.json' \
		--event-loader='postgres://webhookdb:webhookdb@localhost:18005/webhookdb' \
		--event-loader-arg='SELECT request_path as path, request_method as method, request_headers as headers, request_body::jsonb as body FROM logged_webhooks WHERE truncated_at IS NULL LIMIT 10'