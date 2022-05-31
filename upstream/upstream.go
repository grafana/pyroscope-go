package upstream

import (
	"time"
)

type Format string

const FormatPprof Format = "pprof"

type Upstream interface {
	Upload(*UploadJob)
}

type UploadJob struct {
	Name            string
	StartTime       time.Time
	EndTime         time.Time
	SpyName         string
	SampleRate      uint32
	Units           string
	AggregationType string
	Format          Format
	Profile         []byte
	PrevProfile     []byte
}
