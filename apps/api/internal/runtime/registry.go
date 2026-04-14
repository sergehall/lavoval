package runtime

import "fmt"

type Registry struct {
	executors map[string]Executor
}

func NewRegistry(executors map[string]Executor) *Registry {
	cloned := make(map[string]Executor, len(executors))
	for key, executor := range executors {
		cloned[key] = executor
	}

	return &Registry{executors: cloned}
}

func (r *Registry) Find(entrypoint string) (Executor, error) {
	executor, ok := r.executors[entrypoint]
	if !ok {
		return nil, fmt.Errorf("executor not found for entrypoint %q", entrypoint)
	}

	return executor, nil
}
