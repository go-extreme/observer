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
