package main

import "fmt"
import "sync"

var wg sync.WaitGroup // Step 1

func whatever(action string) {
	//wg.Done() // Step 3
	fmt.Printf("routine %v finished\n", action)
	wg.Done()
}

func main() {
	actions := []string{"first", "second", "third", "first", "second", "third"}

	for i := 0; i < len(actions); i++ {
		wg.Add(1)               // Step 2
		go whatever(actions[i]) // *
	}
	wg.Wait() // Step 4
	fmt.Println("main finished")
}
