package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type LockCounter struct {
	count []uint
	start time.Time
	sync.Mutex
}

func (t *LockCounter) Print() {
	t.Lock()
	defer t.Unlock()
	fmt.Println("LockCounter", t.count[len(t.count)-1], time.Since(t.start)/time.Duration(len(t.count)))
}

func (t *LockCounter) Increment() {
	t.Lock()
	defer t.Unlock()
	t.count = append(t.count, uint(len(t.count)))
}

type ChanCounter struct {
	count []uint
	ch    chan bool
	sync.Mutex
	start time.Time
}

func (t *ChanCounter) Print() {
	fmt.Println("ChanCounter", t.count[len(t.count)-1], time.Since(t.start)/time.Duration(len(t.count)))
}

func (t *ChanCounter) Increment() {
	t.ch <- true
}
func (t *ChanCounter) Worker() {
	go func() {
		for {
			_, ok := <-t.ch
			if !ok {
				return
			}
			t.count = append(t.count, uint(len(t.count)))
		}
	}()
}

const spawnRoutines = 100
const chanSize = 100_000

var counterA *LockCounter
var counterB *ChanCounter

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Invalid len of args")
		os.Exit(1)
	}
	f := os.Args[1]
	switch f {
	case "lock":
		counterA = &LockCounter{
			start: time.Now(),
		}
		lockCounter()
	case "chan":
		counterB = &ChanCounter{
			ch:    make(chan bool, chanSize),
			start: time.Now(),
		}
		chanCounter()
	default:
		fmt.Println("Invalid arg")
		os.Exit(1)
	}
}

var wg sync.WaitGroup

func lockCounter() {
	go func() {
		for {
			time.Sleep(time.Second)
			counterA.Print()
		}
	}()
	for i := 0; i < spawnRoutines; i++ {
		wg.Add(1)
		go func() {
			for {
				counterA.Increment()
			}
		}()
	}
	wg.Wait()
}
func chanCounter() {
	go func() {
		for {
			time.Sleep(time.Second)
			counterB.Print()
		}
	}()
	counterB.Worker()
	for i := 0; i < spawnRoutines; i++ {
		wg.Add(1)
		go func() {
			for {
				counterB.Increment()
			}
		}()
	}
	wg.Wait()
}
