package webm

import (
	"testing"
)

func TestSearch(t *testing.T) {
	st := newSeekIndex()
	st.append(seekEntry{30, 300})
	st.append(seekEntry{10, 100})
	st.append(seekEntry{0, 0})
	st.append(seekEntry{20, 200})
	//	t.Log(st)
	t.Log("0", st.search(5))
	t.Log("0", st.search(0))
	t.Log("20", st.search(25))
	t.Log("20", st.search(20))
	t.Log("30", st.search(30))
	t.Log("30", st.search(35))
}
