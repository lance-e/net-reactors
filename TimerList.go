package netreactors

import (
	"time"
)

type TimeEntry struct {
	when_  time.Time
	timer_ *Timer
}

type TimerList []TimeEntry

// implement the heap.Interface:
// implement the sort.Interface
func (t TimerList) Len() int {
	return len(t)
}

func (t TimerList) Less(i, j int) bool {
	return t[i].when_.Before(t[j].when_)
}

func (t TimerList) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// implement Push and Pop
func (t *TimerList) Push(x any) {
	*t = append(*t, x.(TimeEntry))
}

func (t *TimerList) Pop() any {
	old := *t
	n := len(old)
	x := old[n-1]
	*t = old[:n-1]
	return x
}
