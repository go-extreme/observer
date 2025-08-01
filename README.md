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


## 🧩 HTTP Server Example

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-extreme/observer"
)

type User struct {
	Name string `json:"name"`
}

type UserObserver struct{}

func (UserObserver) Created(u User) {
	fmt.Printf("[Created] User '%s' was created\n", u.Name)
}

func (UserObserver) BeforeDelete(u User) {
	fmt.Printf("[BeforeDelete] User '%s' will be deleted\n", u.Name)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	observer.Notify(observer.EventCreated, user)

	fmt.Fprintf(w, "User '%s' created\n", user.Name)
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	observer.NotifyAsync(observer.EventBeforeDelete, user)

	fmt.Fprintf(w, "User '%s' deleted\n", user.Name)
}

func main() {
	observer.SetDebug(true)

	// Attach observer once globally before handling requests
	observer.Attach(User{}, UserObserver{})

	http.HandleFunc("/user/create", createUserHandler)
	http.HandleFunc("/user/delete", deleteUserHandler)

	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

```


## 🧩 Advanced Example

```go
package main

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-extreme/observer"
)

// User entity
type User struct {
	ID   int
	Name string
}

// Observer for User events
type UserObserver struct{}

func (UserObserver) Created(u User) {
	fmt.Printf("[Observer] User '%s' created\n", u.Name)
}

func (UserObserver) Updated(u User) {
	fmt.Printf("[Observer] User '%s' updated\n", u.Name)
}

func (UserObserver) BeforeDelete(u User) {
	fmt.Printf("[Observer] User '%s' will be deleted\n", u.Name)
}

func (UserObserver) Deleted(u User) {
	fmt.Printf("[Observer] User '%s' deleted\n", u.Name)
}

// UserRepository interface
type UserRepository interface {
	Create(user User) (User, error)
	Update(user User) (User, error)
	Delete(id int) error
	GetByID(id int) (User, error)
}

// InMemoryUserRepository implements UserRepository
type InMemoryUserRepository struct {
	mu     sync.Mutex
	data   map[int]User
	lastID int
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		data: make(map[int]User),
	}
}

func (r *InMemoryUserRepository) Create(user User) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastID++
	user.ID = r.lastID
	r.data[user.ID] = user
	return user, nil
}

func (r *InMemoryUserRepository) Update(user User) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.data[user.ID]; !exists {
		return User{}, errors.New("user not found")
	}
	r.data[user.ID] = user
	return user, nil
}

func (r *InMemoryUserRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.data[id]; !exists {
		return errors.New("user not found")
	}
	delete(r.data, id)
	return nil
}

func (r *InMemoryUserRepository) GetByID(id int) (User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, exists := r.data[id]
	if !exists {
		return User{}, errors.New("user not found")
	}
	return user, nil
}

// UserService - handles business logic & triggers observer events
type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(name string) (User, error) {
	user, err := s.repo.Create(User{Name: name})
	if err != nil {
		return User{}, err
	}
	observer.Notify(observer.EventCreated, user)
	return user, nil
}

func (s *UserService) UpdateUser(user User) (User, error) {
	updated, err := s.repo.Update(user)
	if err != nil {
		return User{}, err
	}
	observer.Notify(observer.EventUpdated, updated)
	return updated, nil
}

func (s *UserService) DeleteUser(id int) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	observer.Notify(observer.EventBeforeDelete, user)
	err = s.repo.Delete(id)
	if err != nil {
		return err
	}
	observer.Notify(observer.EventDeleted, user)
	return nil
}

func main() {
	observer.SetDebug(true)

	repo := NewInMemoryUserRepository()
	service := NewUserService(repo)

	// Attach observer for User
	observer.Attach(User{}, UserObserver{})

	// Create user
	user, _ := service.CreateUser("Alice")

	// Update user
	user.Name = "Alice Smith"
	service.UpdateUser(user)

	// Delete user
	service.DeleteUser(user.ID)
}

```
