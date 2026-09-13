export CGO_ENABLED = 0
export NEXT_TELEMETRY_DISABLED = 1

# All build artifacts land in ./releases at the repo root.
RELEASES := $(CURDIR)/releases

.PHONY: build
build: build-desktop

# build-desktop produces the native desktop app for the host OS. Wails must
# build it (a plain `go build` leaves the WebView2 environment without the
# application manifest and hangs at startup), so this uses the wails CLI with
# the frontend build handled separately by build-admin. Wails writes to its own
# build/bin, so the result is copied into ./releases. The output name is set by
# wails.json (Hettix.exe on Windows, Hettix on Linux, Hettix.app on macOS).
# On Linux, webkit2gtk-4.1 (modern distros) needs the webkit2_41 build tag.
WAILS_TAGS := $(if $(filter Linux,$(shell uname -s)),-tags webkit2_41,)
.PHONY: build-desktop
build-desktop: build-admin
	cd cmd/hettix-desktop && wails build -s -skipbindings $(WAILS_TAGS)
	mkdir -p $(RELEASES)
	cp -R cmd/hettix-desktop/build/bin/. $(RELEASES)/

.PHONY: build-admin
build-admin:
	cd admin && \
	yarn install --frozen-lockfile && \
	yarn run build && \
	rm -rf ../pkg/adminui/admin && \
	cp -R dist ../pkg/adminui/admin

.PHONY: clean
clean:
	rm -rf ./releases
	rm -rf ./pkg/adminui/admin
	rm -rf ./admin/dist
	rm -rf ./admin/.next
	rm -rf ./cmd/hettix-desktop/build/bin