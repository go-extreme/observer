package observer

import (
	"fmt"
	"reflect"1
	"sync"
)

// ObserverEventType represents the event type (e.g. BeforeCreate, AfterDelete).
type ObserverEventType string

// Observables – any struct that can return observers via Observer().
type Observables interface {
	Observer() []any
}

// Observer – an optional interface for defining lifecycle methods (like Created, Updated, etc.).
// This is useful if you want to define contracts for different observer behaviors later.
type Observer interface {
	// This is intentionally left empty for now.
	// etc...
}

var (
	eventRegistryMu sync.RWMutex
	eventRegistry   = make(map[ObserverEventType]struct{})
	debug           = false // default off

)

const (
	// 🔵 CREATE LIFECYCLE
	EventBeforeCreate ObserverEventType = "BeforeCreate" // fires before creation logic starts
	EventOnCreating   ObserverEventType = "OnCreating"   // alias for beforeCreate (semantic)
	EventCreated      ObserverEventType = "Created"      // fires after model is persisted
	EventAfterCreate  ObserverEventType = "AfterCreate"  // alias for created

	// 🟠 UPDATE LIFECYCLE
	EventBeforeUpdate ObserverEventType = "BeforeUpdate"
	EventOnUpdating   ObserverEventType = "OnUpdating"
	EventUpdated      ObserverEventType = "Updated"
	EventAfterUpdate  ObserverEventType = "AfterUpdate"

	// 🔴 DELETE LIFECYCLE
	EventBeforeDelete ObserverEventType = "BeforeDelete"
	EventOnDeleting   ObserverEventType = "OnDeleting"
	EventDeleted      ObserverEventType = "Deleted"
	EventAfterDelete  ObserverEventType = "AfterDelete"

	// 🟢 SAVE LIFECYCLE (generic for both create/update)
	EventBeforeSave ObserverEventType = "BeforeSave"
	EventOnSaving   ObserverEventType = "OnSaving"
	EventSaved      ObserverEventType = "Saved"
	EventAfterSave  ObserverEventType = "AfterSave"

	// ⚪ RESTORE LIFECYCLE (soft delete restore)
	EventBeforeRestore ObserverEventType = "BeforeRestore"
	EventOnRestoring   ObserverEventType = "OnRestoring"
	EventRestored      ObserverEventType = "Restored"
	EventAfterRestore  ObserverEventType = "AfterRestore"
)

// Initialize with your built-in events
func init() {
	// Add built-in events to registry
	for _, ev := range []ObserverEventType{
		EventBeforeCreate, EventOnCreating, EventCreated, EventAfterCreate,
		EventBeforeUpdate, EventOnUpdating, EventUpdated, EventAfterUpdate,
		EventBeforeDelete, EventOnDeleting, EventDeleted, EventAfterDelete,
		EventBeforeSave, EventOnSaving, EventSaved, EventAfterSave,
		EventBeforeRestore, EventOnRestoring, EventRestored, EventAfterRestore,
	} {
		eventRegistry[ev] = struct{}{}
	}
}

type Dispatcher struct {
	mu        sync.RWMutex
	observers map[reflect.Type][]any
}

var globalDispatcher = NewDispatcher()

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		observers: make(map[reflect.Type][]any),
	}
}

// SetDebug enables or disables debug logging dynamically
func SetDebug(enabled bool) {
	debug = enabled
}

// debugPrintf prints formatted string only if debug is enabled
func debugPrintf(format string, a ...any) {
	if debug {
		fmt.Printf(format, a...)
	}
}
func Global() *Dispatcher {
	return globalDispatcher
}

// ✅ registerModel automatically finds Observer() and registers observers
func (d *Dispatcher) registerModel(model any) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	debugPrintf("🔄 Registering model: %s\n", modelType.Name())

	// ✅ Check if model implements Observables
	var instance any
	if reflect.TypeOf(model).Kind() == reflect.Ptr {
		instance = model
	} else {
		instance = reflect.New(modelType).Interface()
	}

	if obsModel, ok := instance.(Observables); ok {
		d.mu.Lock()
		// ✅ If observers already registered for this model, skip
		if _, exists := d.observers[modelType]; exists {
			debugPrintf("⚠️ %s already registered, skipping duplicate registration, dont worry the observer will handle duplicate registrations\n", modelType.Name())
		} else {
			observers := obsModel.Observer()
			d.observers[modelType] = append(d.observers[modelType], observers...)
			debugPrintf("✅ %d observers registered for %s\n", len(observers), modelType.Name())
		}
		d.mu.Unlock()

	} else {
		debugPrintf("⚠️ %s does NOT implement Observables\n", modelType.Name())
	}
}

// ✅ Notify sync event
func (d *Dispatcher) dispatchEvent(event ObserverEventType, model any) {
	modelType := normalizeModelType(model)
	debugPrintf("🚀 Dispatching SYNC event '%s' for %s\n", event, modelType.Name())

	d.mu.RLock()
	observers, ok := d.observers[modelType]
	d.mu.RUnlock()

	if !ok {
		debugPrintf("⚠️ No observers for %s\n", modelType.Name())
		return
	}

	var wg sync.WaitGroup
	for _, obs := range observers {
		method := reflect.ValueOf(obs).MethodByName(string(event))
		if method.IsValid() {
			wg.Add(1)
			callObserverMethod(method, model)
			wg.Done()
		}
	}
	wg.Wait()
}

// ✅ Notify async event (non-blocking)
func (d *Dispatcher) dispatchEventAsync(event ObserverEventType, model any) {
	modelType := normalizeModelType(model)
	debugPrintf("🚀 Dispatching ASYNC event '%s' for %s\n", event, modelType.Name())

	d.mu.RLock()
	observers, ok := d.observers[modelType]
	d.mu.RUnlock()

	if !ok {
		debugPrintf("⚠️ No observers for %s\n", modelType.Name())
		return
	}

	for _, obs := range observers {
		method := reflect.ValueOf(obs).MethodByName(string(event))
		if method.IsValid() {
			go callObserverMethod(method, model)
		}
	}
}

func callObserverMethod(method reflect.Value, model any) {
	arg := reflect.ValueOf(model)
	if arg.Type() != method.Type().In(0) {
		if arg.Type().Kind() == reflect.Ptr && method.Type().In(0).Kind() != reflect.Ptr {
			arg = arg.Elem() // convert *User → User
		} else if method.Type().In(0).Kind() == reflect.Ptr && arg.Type().Kind() != reflect.Ptr {
			ptr := reflect.New(arg.Type())
			ptr.Elem().Set(arg)
			arg = ptr
		}
	}
	method.Call([]reflect.Value{arg})
}

func normalizeModelType(model any) reflect.Type {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// RegisterEventType adds a new event type to the registry dynamically
func RegisterEventType(event ObserverEventType) {
	eventRegistryMu.Lock()
	defer eventRegistryMu.Unlock()
	eventRegistry[event] = struct{}{}
	debugPrintf("🆕 Registered new event type: %s\n", event)
}

// IsEventTypeRegistered checks if the event type is known/registered
func IsEventTypeRegistered(event ObserverEventType) bool {
	eventRegistryMu.RLock()
	defer eventRegistryMu.RUnlock()
	_, ok := eventRegistry[event]
	return ok
}

// ListRegisteredEvents returns all registered event types
func ListRegisteredEvents() []ObserverEventType {
	eventRegistryMu.RLock()
	defer eventRegistryMu.RUnlock()

	events := make([]ObserverEventType, 0, len(eventRegistry))
	for ev := range eventRegistry {
		events = append(events, ev)
	}
	return events
}

// Attach registers an observer instance for a specific model type, avoiding duplicates.
func (d *Dispatcher) Attach(model any, observer any) {
	modelType := normalizeModelType(model)

	d.mu.Lock()
	defer d.mu.Unlock()

	// check if the observer is already attached
	existing := d.observers[modelType]
	for _, obs := range existing {
		if reflect.TypeOf(obs) == reflect.TypeOf(observer) {
			debugPrintf("⚠️ Observer %T already attached to %s, skipping duplicate\n", observer, modelType.Name())
			return
		}
	}

	// attach the observer
	d.observers[modelType] = append(d.observers[modelType], observer)
	debugPrintf("✅ Observer %T attached to %s\n", observer, modelType.Name())
}

//
// ✅ GLOBAL HELPERS
//

// ✅ Register registers a model globally
func Register(model any) {
	Global().registerModel(model)
}

// ✅ Notify dispatches synchronously
func Notify(event ObserverEventType, model any) {
	Global().dispatchEvent(event, model)
}

// ✅ NotifyAsync dispatches asynchronously
func NotifyAsync(event ObserverEventType, model any) {
	Global().dispatchEventAsync(event, model)
}

// Global helper for Attach
func Attach(model any, observer any) {
	Global().Attach(model, observer)
}
