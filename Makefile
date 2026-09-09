export CGO_ENABLED = 0
export NEXT_TELEMETRY_DISABLED = 1

.PHONY: build
build: build-admin
	go build ./cmd/hettix

# build-desktop produces the native desktop app. Wails must build it (a plain
# `go build` leaves the WebView2 environment without the application manifest
# and hangs at startup), so this uses the wails CLI with the frontend build
# handled separately by build-admin.
.PHONY: build-desktop
build-desktop: build-admin
	cd cmd/hettix-desktop && \
	wails build -s -skipbindings -o Hettix.exe

.PHONY: build-admin
build-admin:
	cd admin && \
	yarn install --frozen-lockfile && \
	yarn run build && \
	rm -rf ../pkg/adminui/admin && \
	cp -R dist ../pkg/adminui/admin

.PHONY: clean
clean:
	rm -f hettix
	rm -rf ./pkg/adminui/admin
	rm -rf ./admin/dist
	rm -rf ./admin/.next
	rm -rf ./cmd/hettix-desktop/build/bin