package updatecheck

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/yarkovaleksei/go-desktop-kit/updater"
)

// fakeClock — часы, которыми управляет тест: цикл сообщает в waits, сколько
// хочет ждать, и ждёт, пока тест не отправит tick.
type fakeClock struct {
	waits chan time.Duration
	ticks chan time.Time
}

func newClock() *fakeClock {
	return &fakeClock{waits: make(chan time.Duration), ticks: make(chan time.Time)}
}

func (c *fakeClock) After(d time.Duration) <-chan time.Time {
	c.waits <- d
	return c.ticks
}

// harness — цикл с подставными часами, проверкой и уведомителем.
type harness struct {
	t     *testing.T
	clock *fakeClock
	loop  *Loop

	mu       sync.Mutex
	statuses []updater.Status // ответы проверки по очереди; пусто — «актуально»
	checks   chan struct{}
	notified []string
}

func newHarness(t *testing.T, format updater.Format) *harness {
	h := &harness{t: t, clock: newClock(), checks: make(chan struct{}, 10)}
	h.loop = New(Options{
		Check:  h.check,
		Notify: h.notify,
		Format: format,
		After:  h.clock.After,
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	t.Cleanup(h.loop.Stop)
	return h
}

func (h *harness) check(context.Context) updater.Status {
	h.mu.Lock()
	st := updater.Status{State: updater.StateCurrent}
	if len(h.statuses) > 0 {
		st, h.statuses = h.statuses[0], h.statuses[1:]
	}
	h.mu.Unlock()
	h.checks <- struct{}{}
	return st
}

func (h *harness) notify(v string) {
	h.mu.Lock()
	h.notified = append(h.notified, v)
	h.mu.Unlock()
}

func (h *harness) answer(sts ...updater.Status) {
	h.mu.Lock()
	h.statuses = append(h.statuses, sts...)
	h.mu.Unlock()
}

// wait ждёт, что цикл запросит таймер на d.
func (h *harness) wait(d time.Duration) {
	h.t.Helper()
	select {
	case got := <-h.clock.waits:
		if got != d {
			h.t.Fatalf("ожидание %v, want %v", got, d)
		}
	case <-time.After(2 * time.Second):
		h.t.Fatalf("цикл не запросил ожидание %v", d)
	}
}

// tick отправляет срабатывание таймера и ждёт проверки.
func (h *harness) tick() {
	h.t.Helper()
	h.clock.ticks <- time.Time{}
	select {
	case <-h.checks:
	case <-time.After(2 * time.Second):
		h.t.Fatal("проверки после срабатывания таймера нет")
	}
}

// quiet убеждается, что цикл больше не просит таймер и не проверяет.
func (h *harness) quiet() {
	h.t.Helper()
	select {
	case d := <-h.clock.waits:
		h.t.Fatalf("лишнее ожидание %v", d)
	case <-h.checks:
		h.t.Fatal("лишняя проверка")
	case <-time.After(50 * time.Millisecond):
	}
}

func (h *harness) notifiedVersions() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string(nil), h.notified...)
}

func available(v string) updater.Status {
	return updater.Status{State: updater.StateAvailable, Latest: v}
}

func TestDisabledNoCalls(t *testing.T) {
	h := newHarness(t, updater.FormatDeb)
	h.loop.Set(false)
	h.quiet()
	if h.loop.Running() {
		t.Error("цикл запущен при выключенной настройке")
	}
}

func TestNoFormatNoLoop(t *testing.T) {
	h := newHarness(t, updater.FormatNone)
	h.loop.Set(true)
	h.quiet()
	if h.loop.Running() {
		t.Error("цикл запущен у сборки без формата")
	}
}

func TestSchedule(t *testing.T) {
	h := newHarness(t, updater.FormatDeb)
	h.loop.Set(true)
	h.wait(DefaultFirstDelay)
	h.tick()
	h.wait(DefaultInterval)
	h.tick()
	h.wait(DefaultInterval)
}

func TestOneNotificationPerVersion(t *testing.T) {
	h := newHarness(t, updater.FormatDeb)
	h.answer(available("1.2.0"), available("1.2.0"), available("1.3.0"))
	h.loop.Set(true)
	h.wait(DefaultFirstDelay)
	for range 3 {
		h.tick()
		h.wait(DefaultInterval)
	}
	if got := h.notifiedVersions(); len(got) != 2 || got[0] != "1.2.0" || got[1] != "1.3.0" {
		t.Errorf("уведомления: %v", got)
	}

	// Перезапуск цикла не повторяет уведомление о той же версии.
	h.loop.Stop()
	h.answer(available("1.3.0"))
	h.loop.Set(true)
	h.wait(DefaultFirstDelay)
	h.tick()
	h.wait(DefaultInterval)
	if got := h.notifiedVersions(); len(got) != 2 {
		t.Errorf("после перезапуска: %v", got)
	}
}

func TestCurrentAndErrorDoNotNotify(t *testing.T) {
	h := newHarness(t, updater.FormatDeb)
	h.answer(
		updater.Status{State: updater.StateCurrent},
		updater.Status{State: updater.StateFailed, Error: &updater.Failure{Code: "update.network"}},
		updater.Status{State: updater.StateAvailable}, // без версии — не о чем сообщать
	)
	h.loop.Set(true)
	h.wait(DefaultFirstDelay)
	for range 3 {
		h.tick()
		h.wait(DefaultInterval)
	}
	if got := h.notifiedVersions(); len(got) != 0 {
		t.Errorf("уведомления: %v", got)
	}
}

func TestStop(t *testing.T) {
	h := newHarness(t, updater.FormatDeb)
	h.loop.Set(true)
	h.wait(DefaultFirstDelay)
	h.loop.Stop()
	if h.loop.Running() {
		t.Error("цикл запущен после Stop")
	}
	h.quiet()
	h.loop.Stop() // повторный Stop безопасен
}

func TestReenableSingleLoop(t *testing.T) {
	h := newHarness(t, updater.FormatDeb)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.loop.Set(true)
		}()
	}
	wg.Wait()
	h.wait(DefaultFirstDelay)
	h.quiet() // второго цикла, который тоже ждал бы таймер, нет

	h.loop.Set(false)
	h.loop.Set(true)
	h.wait(DefaultFirstDelay)
	h.quiet()
}
