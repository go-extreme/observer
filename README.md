# observer

A lightweight, concurrent-safe **Observer Pattern** implementation for Go.  
Observer helps you build **event-driven** and **decoupled** systems by letting observers (subscribers) react to state changes in subjects (publishers).

---

## 🚀 Features

- 🔄 **Simple API** — Attach, Detach, Notify observers with ease.  
- 🚀 **Dual Notification Modes** — `Fire()` waits for observers; `FireAsync()` doesn’t.  
- 🛡️ **Concurrency-Safe** — No race conditions when adding or notifying observers.  
- ✅ **Duplicate Prevention** — An observer won’t be attached more than once.  
- 🧩 **Lightweight & Extensible** — Perfect for logging, events, or plugin systems.

---

## 📦 Installation

```bash
go get github.com/go-extreme/observer

```

## 🧩  Example Usage

```go
package main

import (
	"fmt"
	"github.com/go-extreme/observer"
)

// User model implements Observables interface by defining Observer() method
type User struct {
	Name string
}

// Observer returns list of observers attached to User
func (u User) Observer() []any {
	return []any{UserObserver{}}
}

// UserObserver implements event handler methods for User lifecycle events
type UserObserver struct{}

// Created event handler (called after user creation)
func (UserObserver) Created(u User) {
	fmt.Printf("[Created] User '%s' was created\n", u.Name)
}

// BeforeDelete event handler (called before user deletion)
func (UserObserver) BeforeDelete(u User) {
	fmt.Printf("[BeforeDelete] User '%s' will be deleted\n", u.Name)
}

func main() {
	// Enable debug logging (optional)
	observer.SetDebug(true)

	// Register User model globally (auto-registers its observers)
	observer.Register(User{})

	user := User{Name: "John"}

	// Synchronous notification - waits until observers finish
	observer.Notify(observer.EventCreated, user)

	// Asynchronous notification - returns immediately, handlers run in goroutines
	observer.NotifyAsync(observer.EventBeforeDelete, user)
}
```
