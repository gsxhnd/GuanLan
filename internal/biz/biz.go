package biz

import "github.com/gsxhnd/guanlan/internal/data"

// Portfolio owns portfolio domain calculations over a Store.
type Portfolio struct {
	Store *data.Store
}

// Task owns sync-task domain policy over a Store.
type Task struct {
	Store *data.Store
}

// Watchlist owns watchlist domain policy over a Store.
type Watchlist struct {
	Store *data.Store
	Tasks *Task
}

// Services groups domain use-cases for HTTP handlers.
type Services struct {
	Portfolio *Portfolio
	Task      *Task
	Watchlist *Watchlist
	Store     *data.Store
}

// New builds biz services from a Store.
func New(store *data.Store) *Services {
	tasks := &Task{Store: store}
	return &Services{
		Portfolio: &Portfolio{Store: store},
		Task:      tasks,
		Watchlist: &Watchlist{Store: store, Tasks: tasks},
		Store:     store,
	}
}
