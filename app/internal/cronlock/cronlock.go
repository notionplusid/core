package cronlock

import "sync"

// Lock meant to be used for cron http-based jobs to prevent racing executions.
type Lock struct {
	locked bool
	mu     *sync.Mutex
}

// New is a Lock constructor.
func New() *Lock {
	return &Lock{mu: &sync.Mutex{}}
}

// ChangeState of the Lock.
// Returns true in case of successful lock state change.
// Returns false in case of an attempt to change state true in case if it's already true.
func (cl *Lock) ChangeState(l bool) bool {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if l && cl.locked {
		return false
	}

	cl.locked = l

	return true
}
