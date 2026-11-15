// Package updatecheck — фоновая проверка обновлений по настройке
// «Проверять обновления автоматически» (D-031).
//
// Цикл проверяет обновления через FirstDelay после включения и затем раз в
// Interval. Если вышла новая версия, он один раз на версию вызывает Notify.
// Включение и выключение настройки запускают и останавливают цикл без
// перезапуска приложения. Сборка без формата пакета обновляться из
// приложения не умеет, и цикл для неё не запускается.
//
// Пакет не знает о Wails: проверку и уведомление подставляет main.go.
package updatecheck

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/yarkovaleksei/go-desktop-kit/updater"
)

// Расписание по умолчанию: первая проверка через минуту после включения —
// запуск приложения не ждёт сети, — затем раз в сутки.
const (
	DefaultFirstDelay = time.Minute
	DefaultInterval   = 24 * time.Hour
)

// Options — зависимости цикла.
type Options struct {
	// Check проверяет выпуски; ошибка сети приходит в Status.Error.
	Check func(ctx context.Context) updater.Status
	// Notify сообщает о новой версии. Вызывается один раз на версию.
	Notify func(version string)
	// Format — формат пакета этой сборки; FormatNone — цикл не запускается.
	Format updater.Format

	FirstDelay time.Duration // 0 — DefaultFirstDelay
	Interval   time.Duration // 0 — DefaultInterval
	// After — таймер; nil — time.After. Тесты подставляют свои часы.
	After func(time.Duration) <-chan time.Time
	Log   *slog.Logger
}

// Loop — фоновый цикл проверки. Методы безопасны для одновременного вызова.
type Loop struct {
	o Options

	mu       sync.Mutex
	cancel   context.CancelFunc // nil — цикл не запущен
	done     chan struct{}
	notified string // версия, о которой уже сообщили
}

// New готовит цикл; он не запущен, пока не вызван Set(true).
func New(o Options) *Loop {
	if o.FirstDelay <= 0 {
		o.FirstDelay = DefaultFirstDelay
	}
	if o.Interval <= 0 {
		o.Interval = DefaultInterval
	}
	if o.After == nil {
		o.After = time.After
	}
	if o.Log == nil {
		o.Log = slog.Default()
	}
	return &Loop{o: o}
}

// Set запускает цикл (on) или останавливает его. Повторный Set с тем же
// значением ничего не меняет: второго цикла не бывает.
func (l *Loop) Set(on bool) {
	if !on {
		l.Stop()
		return
	}
	if l.o.Format == updater.FormatNone {
		l.o.Log.Info("фоновая проверка обновлений не запущена: сборка без формата пакета")
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	l.cancel, l.done = cancel, make(chan struct{})
	go l.run(ctx, l.done)
	l.o.Log.Info("фоновая проверка обновлений включена")
}

// Stop останавливает цикл и ждёт его завершения: идущая проверка
// прерывается. Без запущенного цикла ничего не делает.
func (l *Loop) Stop() {
	l.mu.Lock()
	cancel, done := l.cancel, l.done
	l.cancel, l.done = nil, nil
	l.mu.Unlock()
	if cancel == nil {
		return
	}
	cancel()
	<-done
	l.o.Log.Info("фоновая проверка обновлений выключена")
}

// Running сообщает, запущен ли цикл.
func (l *Loop) Running() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cancel != nil
}

func (l *Loop) run(ctx context.Context, done chan struct{}) {
	defer close(done)
	wait := l.o.FirstDelay
	for {
		select {
		case <-ctx.Done():
			return
		case <-l.o.After(wait):
		}
		l.checkOnce(ctx)
		wait = l.o.Interval
	}
}

// checkOnce проверяет обновления и сообщает о новой версии, если о ней ещё
// не сообщали.
func (l *Loop) checkOnce(ctx context.Context) {
	st := l.o.Check(ctx)
	if ctx.Err() != nil {
		return
	}
	if st.State != updater.StateAvailable || st.Latest == "" {
		if st.Error != nil {
			l.o.Log.Warn("фоновая проверка обновлений не удалась", "code", st.Error.Code)
		}
		return
	}
	l.mu.Lock()
	fresh := l.notified != st.Latest
	l.notified = st.Latest
	l.mu.Unlock()
	if fresh {
		l.o.Log.Info("доступна новая версия", "version", st.Latest)
		l.o.Notify(st.Latest)
	}
}
