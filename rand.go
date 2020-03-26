// Go's _select_ lets you wait on multiple channel
// operations. Combining goroutines and channels with
// select is a powerful feature of Go.

package main

import "time"
import "math/rand"
import "fmt"

type Duration int64

func main() {
    fmt.Println("LOL")
    for i:=0; i<5; i++ {
    	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	fmt.Println(".")
    }
    fmt.Println("LOOL")
}

