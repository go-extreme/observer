package observer

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

//
// ──────────────────────────────
//   1️⃣ BASIC FUNCTIONALITY TEST
// ──────────────────────────────
//

// TestLogger is a dummy observer for testing
type TestLogger struct {
	createdCount int32
	updatedCount int32
}

func (t *TestLogger) Created(u TestUser) {
	atomic.AddInt32(&t.createdCount, 1)
}

func (t *TestLogger) Updated(u TestUser) {
	atomic.AddInt32(&t.updatedCount, 1)
}

// TestUser implements Observables
type TestUser struct {
	ID   int
	Name string
}

func (TestUser) Observer() []any {
	return []any{&testLogger} // use global instance so we can count
}

var testLogger TestLogger

func TestBasicNotify(t *testing.T) {
	SetDebug(true)
	Register(TestUser{})

	u := TestUser{ID: 1, Name: "John"}
	Notify(EventCreated, u)
	Notify(EventUpdated, u)

	if atomic.LoadInt32(&testLogger.createdCount) != 1 {
		t.Errorf("Expected Created() to be called 1 time, got %d", testLogger.createdCount)
	}
	if atomic.LoadInt32(&testLogger.updatedCount) != 1 {
		t.Errorf("Expected Updated() to be called 1 time, got %d", testLogger.updatedCount)
	}
}

//
// ──────────────────────────────
//   2️⃣ DUPLICATE REGISTRATION
// ──────────────────────────────
//

func TestNoDuplicateRegistration(t *testing.T) {
	SetDebug(true)
	Register(TestUser{})
	Register(TestUser{}) // second register should NOT double register

	u := TestUser{ID: 1, Name: "John"}
	atomic.StoreInt32(&testLogger.createdCount, 0)

	Notify(EventCreated, u)

	if atomic.LoadInt32(&testLogger.createdCount) != 1 {
		t.Errorf("Expected Created() to be called once (no duplicates), got %d", testLogger.createdCount)
	}
}

//
// ──────────────────────────────
//   3️⃣ CONCURRENT SAFETY TEST
// ──────────────────────────────
//

func TestConcurrentNotify(t *testing.T) {
	SetDebug(true)

	Register(TestUser{})

	const goroutines = 1000
	atomic.StoreInt32(&testLogger.createdCount, 0)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			Notify(EventCreated, TestUser{ID: 1})
		}()
	}
	wg.Wait()

	if count := atomic.LoadInt32(&testLogger.createdCount); count != goroutines {
		t.Errorf("Expected Created() to be called %d times, got %d", goroutines, count)
	}
}

//
// ──────────────────────────────
//   4️⃣ DYNAMIC EVENT TYPE TEST
// ──────────────────────────────
//

type DynamicObserver struct {
	called int32
}

func (d *DynamicObserver) CustomEvent(u DynamicUser) {
	atomic.AddInt32(&d.called, 1)
}

type DynamicUser struct{}

func (DynamicUser) Observer() []any {
	return []any{&dynamicObs}
}

var dynamicObs DynamicObserver

func TestDynamicEventType(t *testing.T) {
	SetDebug(true)
	var custom ObserverEventType = "CustomEvent"
	RegisterEventType(custom)
	Register(DynamicUser{})

	result := IsEventTypeRegistered(custom)

	fmt.Printf("##### registerd new event: %v\n", result)
	ListRegisteredEvents()
	atomic.StoreInt32(&dynamicObs.called, 0)
	Notify(custom, DynamicUser{})

	if count := atomic.LoadInt32(&dynamicObs.called); count != 1 {
		t.Errorf("Expected CustomEvent() to be called 1 time, got %d", count)
	}
}

//
// ──────────────────────────────
//   5️⃣ BENCHMARKS
// ──────────────────────────────
//

func BenchmarkNotify(b *testing.B) {
	//SetDebug(true)

	Register(TestUser{})

	u := TestUser{ID: 1, Name: "Benchmark"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Notify(EventCreated, u)
	}
}

func BenchmarkNotifyAsync(b *testing.B) {
	//SetDebug(true)

	Register(TestUser{})

	u := TestUser{ID: 1, Name: "Benchmark"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NotifyAsync(EventCreated, u)
	}
}
func BenchmarkNotifyParallel(b *testing.B) {
	Register(TestUser{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Notify(EventCreated, TestUser{ID: 1})
		}
	})
}
func BenchmarkNotifyAsyncParallel(b *testing.B) {
	// Use the global dispatcher for consistency
	Register(TestUser{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NotifyAsync(EventCreated, TestUser{ID: 1})
		}
	})
}
