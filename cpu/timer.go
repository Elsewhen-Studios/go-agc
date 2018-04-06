package cpu

type timer struct {
	n string
	v int
	i int
}

func NewTimer(name string, interval, offset int) *timer {
	return &timer{
		n: name,
		v: offset,
		i: interval,
	}
}

// Inc increments the timer's value and returns true if
// it wrapped around.
func (t *timer) Inc(d int) (wrapped bool) {
	t.v += d
	if t.v >= t.i {
		wrapped = true
		t.v -= t.i
	}
	return
}
