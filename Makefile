export CGO_ENABLED = 0
export NEXT_TELEMETRY_DISABLED = 1

.PHONY: build
build: build-admin
	go build ./cmd/hettix

.PHONY: build-admin
build-admin:
	cd admin && \
	yarn install --frozen-lockfile && \
	yarn run build && \
	rm -rf ../cmd/hettix/admin && \
	cp -R dist ../cmd/hettix/admin

.PHONY: clean
clean:
	rm -f hettix
	rm -rf ./cmd/hettix/admin
	rm -rf ./admin/dist
	rm -rf ./admin/.next