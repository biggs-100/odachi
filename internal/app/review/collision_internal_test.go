package review

import "testing"

// TestCollisionAsksWhatTheRunLanded is the test the review of 0137 asked for. It
// fails if the guard goes back to asking the shape of the candidate -a `repeat`
// flag set from `unit.of > 1`, or any other property of the diff- instead of what
// this run already stored: a part of a split candidate may merge into a lens it
// landed itself, and a run that landed nothing may not merge at all.
func TestCollisionAsksWhatTheRunLanded(t *testing.T) {
	cases := []struct {
		name                     string
		sealed, existing, landed []string
		want                     string
	}{
		{
			name:   "a store that holds nothing yet",
			sealed: []string{"reliability"}, existing: nil, landed: nil,
			want: "",
		},
		{
			name:   "a lens nobody answered",
			sealed: []string{"reliability"}, existing: []string{"risk"}, landed: nil,
			want: "",
		},
		{
			name:   "a second run over a lens the store holds",
			sealed: []string{"reliability"}, existing: []string{"reliability"}, landed: nil,
			want: "reliability",
		},
		{
			name:   "a part of the same run, answering the lens it landed",
			sealed: []string{"reliability"}, existing: []string{"reliability"}, landed: []string{"reliability"},
			want: "",
		},
		{
			name:   "a part that landed one lens and is handed another the store holds",
			sealed: []string{"risk", "reliability"}, existing: []string{"reliability"}, landed: []string{"risk"},
			want: "reliability",
		},
		{
			name:   "the same lens spelled another way",
			sealed: []string{"Reliability"}, existing: []string{"reliability"}, landed: nil,
			want: "Reliability",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := collision(c.sealed, c.existing, c.landed); got != c.want {
				t.Fatalf("collision = %q, want %q", got, c.want)
			}
		})
	}
}
