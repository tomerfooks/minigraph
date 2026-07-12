package minigraph

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"sync"
)

// Checkpointer persists the latest Step of each thread, making runs durable:
// a run that crashed, was interrupted, or was broken out of continues from
// its last saved Step. Implementations should serialize S (e.g. JSON) when
// storing outside process memory.
type Checkpointer[S any] interface {
	Save(ctx context.Context, thread string, step Step[S]) error
	Load(ctx context.Context, thread string) (Step[S], bool, error)
}

// MemorySaver is an in-process Checkpointer, good for tests and single-run
// durability (interrupt/resume). The zero value is ready to use. Steps are
// stored by value: reference fields inside S still point at shared data.
type MemorySaver[S any] struct {
	mu      sync.Mutex
	threads map[string]Step[S]
}

func (m *MemorySaver[S]) Save(_ context.Context, thread string, step Step[S]) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.threads == nil {
		m.threads = map[string]Step[S]{}
	}
	m.threads[thread] = step
	return nil
}

func (m *MemorySaver[S]) Load(_ context.Context, thread string) (Step[S], bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	step, ok := m.threads[thread]
	return step, ok, nil
}

// StreamThread is Stream with durability. If the thread has a saved Step the
// run resumes from it and initial is ignored; otherwise the run starts fresh.
// Every successful Step is saved before it is yielded, and an Interrupt's
// Step is saved too — so a process can die at any point and the thread
// continues where it left off. Ordinary node errors are not saved: rerunning
// the thread retries from the last good Step.
func (r *App[S]) StreamThread(ctx context.Context, saver Checkpointer[S], thread string, initial S) iter.Seq2[Step[S], error] {
	return func(yield func(Step[S], error) bool) {
		from := Step[S]{Node: Start, State: initial}
		if cp, ok, err := saver.Load(ctx, thread); err != nil {
			yield(from, fmt.Errorf("load thread %q: %w", thread, err))
			return
		} else if ok {
			from = cp
		}
		for step, err := range r.StreamFrom(ctx, from) {
			if err == nil || errors.As(err, new(*Interrupt)) {
				if serr := saver.Save(ctx, thread, step); serr != nil {
					yield(step, fmt.Errorf("save thread %q: %w", thread, serr))
					return
				}
			}
			if !yield(step, err) || err != nil {
				return
			}
		}
	}
}

// InvokeThread runs a thread to completion (or its next Interrupt) and
// returns the final state. Calling it again on a finished thread is a no-op
// that returns the same final state. To answer an Interrupt, edit the
// returned state, Save it under the interrupted node, and call InvokeThread
// again:
//
//	saver.Save(ctx, thread, Step[S]{Node: intr.Node, State: edited})
func (r *App[S]) InvokeThread(ctx context.Context, saver Checkpointer[S], thread string, initial S) (S, error) {
	final := initial
	if cp, ok, err := saver.Load(ctx, thread); err == nil && ok {
		final = cp.State // a finished thread yields no steps; its state is the answer
	}
	for step, err := range r.StreamThread(ctx, saver, thread, initial) {
		final = step.State
		if err != nil {
			return final, err
		}
	}
	return final, nil
}
