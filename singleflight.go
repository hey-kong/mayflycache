package main

import (
	"sync"
)

type Once struct {
	mu sync.Mutex
	m  map[string]*call
}

// A call is used to handle the function call corresponding to the string in Once.
type call struct {
	wg  sync.WaitGroup // there may be multiple function calls waiting for the same result
	val interface{}    // the value returned by the function call
	err error
}

func (o *Once) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	o.mu.Lock()
	if o.m == nil {
		o.m = make(map[string]*call)
	}

	// Judge if function call with the key has occurred
	if c, ok := o.m[key]; ok {
		o.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	// Here is the first call
	c := new(call)
	c.wg.Add(1)
	o.m[key] = c
	o.mu.Unlock()

	// Call and get the value, and notify the waiting calls
	c.val, c.err = fn()
	c.wg.Done()

	// Remove the result of this call
	o.mu.Lock()
	delete(o.m, key)
	o.mu.Unlock()

	return c.val, c.err
}
