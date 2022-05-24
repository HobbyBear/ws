package trylock

import (
	"fmt"
	"testing"
)

func TestMutex_TryLock(t *testing.T) {
	l := NewMutex()
	fmt.Println(l.TryLock())
	fmt.Println(l.TryLock())
	l.Unlock()
	fmt.Println(l.TryLock())
}
