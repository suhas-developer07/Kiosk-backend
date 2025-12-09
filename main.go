package main

import (
	"fmt"
	"time"
)

// func main() {
//     msgChan := make(chan string, 2)

//     // Start receiver goroutine BEFORE blocking send

//     // Send 1
//     msgChan <- "Msg-A"
//     fmt.Printf(" M: Sent 'Msg-A'. Len: %d\n", len(msgChan))

//     // Send 2
//     msgChan <- "Msg-B"
//     fmt.Printf(" M: Sent 'Msg-B'. Len: %d (FULL)\n", len(msgChan))

// 	 go func() {
//        time.Sleep(50* time.Millisecond)
// 	   fmt.Println("time is over")
//         val := <-msgChan
//         fmt.Printf(" R: Received '%s'\n", val)
//     }()
//     // Send 3 (will block until goroutine receives)
//     fmt.Println(" M: Attempting to send 'Msg-C'...")
//     msgChan <- "Msg-C"
//     fmt.Println(" M: Sent 'Msg-C' (unblocked).",len(msgChan))

//     time.Sleep(200 * time.Millisecond) // keep main alive
// }

func main() {
	eventA := make(chan string) // Unbuffered
	eventB := make(chan string) // Unbuffered

	// Sender A (Fast)
	go func() { time.Sleep(50 * time.Millisecond); eventA <- "Fast Event" }()

	// Sender B (Slow)
	go func() { time.Sleep(50 * time.Millisecond); eventB <- "Slow Event" }()

	fmt.Println(" M: Waiting for the FIRST event...")

	select {
	case msgA := <-eventA: // Ready at ~50ms
		fmt.Printf(" M: Received: %s\n", msgA)
	case msgB := <-eventB: // Ready at ~200ms
		fmt.Printf(" M: Received: %s\n", msgB)
	}

}
