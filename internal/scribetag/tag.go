package scribetag

import "time"

type ScribeTag struct {
	Name      string // should equal to scribe.Scribe.CanonicalName()
	LastPulse time.Time
	before    string
	after     string
}
