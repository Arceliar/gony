package main

import (
	"fmt"

	"github.com/Arceliar/phony"
)

// Structs can embed the Inbox type to fulfill the Actor interface.
type printer struct {
	phony.Inbox
}

// Functions can be defined to send messages to an Actor from another Actor.
func (p *printer) Println(from phony.Actor, msg ...interface{}) {
	p.RecvFrom(from, func() { fmt.Println(msg...) })
}

// It's useful to embed an Actor in a struct whose fields the Actor is responsible for.
type counter struct {
	phony.Inbox
	count   int
	printer *printer
}

// An EnqueueFrom nil is useful for asking an Actor to do something from non-Actor code.
func (c *counter) Increment() {
	c.RecvFrom(nil, func() { c.count++ })
}

// A SyncExec function returns a channel that will be closed after the message has been processed from the Inbox.
// This can be used to interrogate an Actor from an outside goroutine.
// Note that Actors shouldn't use this on eachother, since it blocks, it's just meant for convenience when interfacing with outside code -- the Actor interface explicitly doesn't include it to make this slightly harder to do by mistake.
func (c *counter) Get(n *int) {
	<-c.SyncExec(func() { *n = c.count })
}

// Print sends a message to the counter, telling to to call c.printer.Println
// Calling Println sends a message to the printer, telling it to print
// So message sends become function calls.
func (c *counter) Print() {
	c.RecvFrom(c, func() {
		c.printer.Println(c, "The count is:", c.count)
	})
}

func main() {
	c := &counter{printer: new(printer)} // Create an actor
	for idx := 0; idx < 10; idx++ {
		c.Increment() // Ask the Actor to do some work
		c.Print()     // And ask it to send a message to another Actor, which handles them asynchronously
	}
	var n int
	c.Get(&n)                         // Inspect the Actor's internal state
	fmt.Println("Value from Get:", n) // This likely prints before the Print() lines above have finished -- Actors work asynchronously.
	<-c.printer.SyncExec(func() {})   // Wait for an Actor to handle a message, in this case just to finish printing
	fmt.Println("Exiting")
}
