package wsintercept_test

import (
	"context"
	"testing"
	"time"

	"github.com/Vibe-Coding-Base/Hettix/pkg/httpql"
	"github.com/Vibe-Coding-Base/Hettix/pkg/wsintercept"
	"github.com/Vibe-Coding-Base/Hettix/pkg/wsproxy"
)

type result struct {
	payload []byte
	drop    bool
}

func message(payload string) wsproxy.Message {
	return wsproxy.Message{ConnID: "conn", Direction: wsproxy.ClientToServer, Opcode: 1, Payload: []byte(payload)}
}

// interceptAsync runs Intercept in a goroutine and returns a channel with its
// result plus a function that waits for the frame to be queued.
func interceptAsync(t *testing.T, svc *wsintercept.Service, ctx context.Context, m wsproxy.Message) chan result {
	t.Helper()

	out := make(chan result, 1)
	go func() {
		payload, drop := svc.Intercept(ctx, m)
		out <- result{payload: payload, drop: drop}
	}()

	return out
}

func waitForFrame(t *testing.T, svc *wsintercept.Service) wsintercept.Frame {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if frames := svc.Frames(); len(frames) == 1 {
			return frames[0]
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("frame was never queued")

	return wsintercept.Frame{}
}

func TestModifyFrame(t *testing.T) {
	svc := wsintercept.NewService(wsintercept.Config{Enabled: true})

	out := interceptAsync(t, svc, context.Background(), message("hello"))
	frame := waitForFrame(t, svc)

	if err := svc.ModifyFrame(frame.ID, []byte("changed")); err != nil {
		t.Fatalf("modify: %v", err)
	}

	got := <-out
	if got.drop || string(got.payload) != "changed" {
		t.Fatalf("expected forwarded payload %q, got %q (drop %v)", "changed", got.payload, got.drop)
	}
}

func TestForwardFrame(t *testing.T) {
	svc := wsintercept.NewService(wsintercept.Config{Enabled: true})

	out := interceptAsync(t, svc, context.Background(), message("hello"))
	frame := waitForFrame(t, svc)

	if err := svc.ForwardFrame(frame.ID); err != nil {
		t.Fatalf("forward: %v", err)
	}

	got := <-out
	if got.drop || string(got.payload) != "hello" {
		t.Fatalf("expected unchanged payload, got %q (drop %v)", got.payload, got.drop)
	}
}

func TestDropFrame(t *testing.T) {
	svc := wsintercept.NewService(wsintercept.Config{Enabled: true})

	out := interceptAsync(t, svc, context.Background(), message("hello"))
	frame := waitForFrame(t, svc)

	if err := svc.DropFrame(frame.ID); err != nil {
		t.Fatalf("drop: %v", err)
	}

	got := <-out
	if !got.drop {
		t.Fatalf("expected the frame to be dropped, got payload %q", got.payload)
	}
}

func TestDisabledBypasses(t *testing.T) {
	svc := wsintercept.NewService(wsintercept.Config{Enabled: false})

	payload, drop := svc.Intercept(context.Background(), message("hello"))
	if drop || string(payload) != "hello" {
		t.Fatalf("disabled interception should forward as-is, got %q (drop %v)", payload, drop)
	}
	if len(svc.Frames()) != 0 {
		t.Fatal("disabled interception should not queue frames")
	}
}

func TestFilterOnlyHoldsMatches(t *testing.T) {
	filter, err := httpql.Parse(`ws.payload cont "secret"`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	svc := wsintercept.NewService(wsintercept.Config{Enabled: true, Filter: filter})

	// Non-matching message is forwarded without being held.
	payload, drop := svc.Intercept(context.Background(), message("harmless"))
	if drop || string(payload) != "harmless" {
		t.Fatalf("non-matching message should pass through, got %q (drop %v)", payload, drop)
	}

	// Matching message is held.
	out := interceptAsync(t, svc, context.Background(), message("this is secret"))
	frame := waitForFrame(t, svc)
	if err := svc.ForwardFrame(frame.ID); err != nil {
		t.Fatalf("forward: %v", err)
	}
	<-out
}

func TestContextCancelForwards(t *testing.T) {
	svc := wsintercept.NewService(wsintercept.Config{Enabled: true})

	ctx, cancel := context.WithCancel(context.Background())
	out := interceptAsync(t, svc, ctx, message("hello"))
	waitForFrame(t, svc)

	cancel()

	got := <-out
	if got.drop || string(got.payload) != "hello" {
		t.Fatalf("teardown should forward as-is, got %q (drop %v)", got.payload, got.drop)
	}
}

func TestUpdateSettingsReleasesFrames(t *testing.T) {
	svc := wsintercept.NewService(wsintercept.Config{Enabled: true})

	out := interceptAsync(t, svc, context.Background(), message("hello"))
	waitForFrame(t, svc)

	// Disabling interception forwards any held frames.
	svc.UpdateSettings(wsintercept.Settings{Enabled: false})

	got := <-out
	if got.drop || string(got.payload) != "hello" {
		t.Fatalf("disabling should forward held frames, got %q (drop %v)", got.payload, got.drop)
	}
}
