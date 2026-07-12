package tomergraph

import "fmt"

// Interrupt pauses a run for outside input — human approval, missing data,
// anything the graph can't decide alone. Return one from a node:
//
//	return s, &Interrupt{Payload: "ok to send this email?"}
//
// Unlike an ordinary error, the state returned alongside an Interrupt is
// kept: the run yields it as the final Step, which is the checkpoint to
// continue from once the outside world has answered (usually after editing
// the state):
//
//	var intr *Interrupt
//	if errors.As(err, &intr) {
//	    state.Approved = true
//	    final, err = app.InvokeFrom(ctx, Step[S]{Node: intr.Node, State: state})
//	}
//
// InvokeFrom routes onward from the interrupted node; the node itself does not
// re-run. Inside Parallel branches an Interrupt cannot pause the run and is
// treated as a plain error.
type Interrupt struct {
	Payload any    // what the node wants to tell whoever resumes the run
	Node    string // set by the engine: the node that paused
}

func (i *Interrupt) Error() string {
	return fmt.Sprintf("tomergraph: interrupted at %q: %v", i.Node, i.Payload)
}
