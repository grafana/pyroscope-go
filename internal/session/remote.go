package session

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pyroscope-io/client/internal/types"
)

type UploadFormat string
type Payload interface {
	Bytes() []byte
}

const (
	Pprof UploadFormat = "pprof"
	Trie               = "trie"
)

type UploadJob struct {
	Name            string
	StartTime       time.Time
	EndTime         time.Time
	SpyName         string
	SampleRate      uint32
	Units           string
	AggregationType string
	Format          UploadFormat
	Profile         []byte
	PrevProfile     []byte
}

type Upstream interface {
	Stop()
	Upload(u *UploadJob)
}

var (
	ErrCloudTokenRequired = errors.New("Please provide an authentication token. You can find it here: https://pyroscope.io/cloud")
	ErrUpload             = errors.New("Failed to upload a profile")
	cloudHostnameSuffix   = "pyroscope.cloud"
)

type Remote struct {
	cfg    RemoteConfig
	jobs   chan *UploadJob
	client *http.Client
	Logger types.Logger

	done chan struct{}
	wg   sync.WaitGroup
}

type RemoteConfig struct {
	AuthToken              string
	UpstreamThreads        int
	UpstreamAddress        string
	UpstreamRequestTimeout time.Duration

	ManualStart bool
}

func NewRemote(cfg RemoteConfig, logger types.Logger) (*Remote, error) {
	remote := &Remote{
		cfg:  cfg,
		jobs: make(chan *UploadJob, 20),
		client: &http.Client{
			Transport: &http.Transport{
				MaxConnsPerHost: cfg.UpstreamThreads,
			},
			Timeout: cfg.UpstreamRequestTimeout,
		},
		Logger: logger,
		done:   make(chan struct{}),
	}

	// parse the upstream address
	u, err := url.Parse(cfg.UpstreamAddress)
	if err != nil {
		return nil, err
	}

	// authorize the token first
	if cfg.AuthToken == "" && requiresAuthToken(u) {
		return nil, ErrCloudTokenRequired
	}

	if !cfg.ManualStart {
		// start goroutines for uploading profile data
		remote.Start()
	}

	return remote, nil
}

func (r *Remote) Start() {
	for i := 0; i < r.cfg.UpstreamThreads; i++ {
		go r.handleJobs()
	}
}

func (r *Remote) Stop() {
	if r.done != nil {
		close(r.done)
	}

	// wait for uploading goroutines exit
	r.wg.Wait()
}

func (r *Remote) Upload(job *UploadJob) {
	select {
	case r.jobs <- job:
	default:
		r.Logger.Errorf("remote upload queue is full, dropping a profile job")
	}
}

func (r *Remote) uploadProfile(j *UploadJob) error {
	u, err := url.Parse(r.cfg.UpstreamAddress)
	if err != nil {
		return fmt.Errorf("url parse: %v", err)
	}

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("profile", "profile.pprof")
	fw.Write(j.Profile)
	if err != nil {
		return err
	}
	if j.PrevProfile != nil {
		fw, err = writer.CreateFormFile("prev_profile", "profile.pprof")
		fw.Write(j.PrevProfile)
		if err != nil {
			return err
		}
	}
	writer.Close()

	q := u.Query()
	q.Set("name", j.Name)
	// TODO: I think these should be renamed to startTime / endTime
	q.Set("from", strconv.Itoa(int(j.StartTime.Unix())))
	q.Set("until", strconv.Itoa(int(j.EndTime.Unix())))
	q.Set("spyName", j.SpyName)
	q.Set("sampleRate", strconv.Itoa(int(j.SampleRate)))
	q.Set("units", j.Units)
	q.Set("aggregationType", j.AggregationType)

	u.Path = path.Join(u.Path, "/ingest")
	u.RawQuery = q.Encode()

	r.Logger.Debugf("uploading at %s", u.String())
	// new a request for the job
	request, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return fmt.Errorf("new http request: %v", err)
	}
	contentType := writer.FormDataContentType()
	r.Logger.Debugf("content type: %s", contentType)
	request.Header.Set("Content-Type", contentType)
	// request.Header.Set("Content-Type", "binary/octet-stream+"+string(j.Format))

	if r.cfg.AuthToken != "" {
		request.Header.Set("Authorization", "Bearer "+r.cfg.AuthToken)
	}

	// do the request and get the response
	response, err := r.client.Do(request)
	if err != nil {
		return fmt.Errorf("do http request: %v", err)
	}
	defer response.Body.Close()

	// read all the response body
	_, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response body: %v", err)
	}

	if response.StatusCode != 200 {
		return ErrUpload
	}

	return nil
}

// handle the jobs
func (r *Remote) handleJobs() {
	for {
		select {
		case <-r.done:
			return
		case job := <-r.jobs:
			r.safeUpload(job)
		}
	}
}

func requiresAuthToken(u *url.URL) bool {
	return strings.HasSuffix(u.Host, cloudHostnameSuffix)
}

// do safe upload
func (r *Remote) safeUpload(job *UploadJob) {
	defer func() {
		if catch := recover(); catch != nil {
			r.Logger.Errorf("recover stack: %v", debug.Stack())
		}
	}()

	// update the profile data to server
	if err := r.uploadProfile(job); err != nil {
		r.Logger.Errorf("upload profile: %v", err)
	}
}
