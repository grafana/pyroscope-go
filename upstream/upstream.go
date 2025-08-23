package upstream

import (
	"time"
)

type Upstream interface {
	Upload(job *UploadJob)
	Flush()
}

type UploadJob struct {
	Name            string
	StartTime       time.Time
	EndTime         time.Time
	SpyName         string
	SampleRate      uint32
	Units           string
	AggregationType string
	Profile         []byte
}
