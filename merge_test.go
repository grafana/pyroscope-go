package client

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/google/pprof/profile"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readCorpusItemFile(fname string, doDecompress bool) *testcase {
	bs, err := ioutil.ReadFile(fname)
	if err != nil {
		panic(err)
	}
	r, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(bs)))
	if err != nil {
		panic(err)
	}
	contentType := r.Header.Get("Content-Type")
	rawData, _ := ioutil.ReadAll(r.Body)
	decompress := func(b []byte) []byte {
		if len(b) < 2 {
			return b
		}
		if b[0] == 0x1f && b[1] == 0x8b {
			gzipr, err := gzip.NewReader(bytes.NewReader(b))
			if err != nil {
				panic(err)
			}
			defer gzipr.Close()
			var buf bytes.Buffer
			if _, err = io.Copy(&buf, gzipr); err != nil {
				panic(err)
			}
			return buf.Bytes()
		}
		return b
	}

	if contentType == "binary/octet-stream" {
		return &testcase{
			profile: decompress(rawData),
			//config:  tree.DefaultSampleTypeMapping,
			fname: fname,
		}
	}
	boundary, err := ParseBoundary(contentType)
	if err != nil {
		panic(err)
	}

	f, err := multipart.NewReader(bytes.NewReader(rawData), boundary).ReadForm(32 << 20)
	if err != nil {
		panic(err)
	}
	const (
		formFieldProfile          = "profile"
		formFieldPreviousProfile  = "prev_profile"
		formFieldSampleTypeConfig = "sample_type_config"
	)

	Profile, err := ReadField(f, formFieldProfile)
	if err != nil {
		panic(err)
	}
	PreviousProfile, err := ReadField(f, formFieldPreviousProfile)
	if err != nil {
		panic(err)
	}

	if doDecompress {
		Profile = decompress(Profile)
		PreviousProfile = decompress(PreviousProfile)
	}
	elem := &testcase{Profile, PreviousProfile, fname, "gospy"}
	return elem
}

type testcase struct {
	profile, prev []byte
	fname         string
	spyname       string
}

func ReadField(form *multipart.Form, name string) ([]byte, error) {
	files, ok := form.File[name]
	if !ok || len(files) == 0 {
		return nil, nil
	}
	fh := files[0]
	if fh.Size == 0 {
		return nil, nil
	}
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		err = f.Close()
	}()
	b := bytes.NewBuffer(make([]byte, 0, fh.Size))
	if _, err = io.Copy(b, f); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func ParseBoundary(contentType string) (string, error) {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}
	boundary, ok := params["boundary"]
	if !ok {
		return "", fmt.Errorf("malformed multipart content type header")
	}
	return boundary, nil
}

func Test(t *testing.T) {
	err := filepath.Walk("/home/korniltsev/trash/qwe",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".txt") {
				return nil
			}

			test := readCorpusItemFile(path, true)
			if test.prev != nil {
				p1, err := profile.Parse(bytes.NewReader(test.prev))
				if err != nil {
					panic(err)
				}
				p2, err := profile.Parse(bytes.NewReader(test.profile))
				if err != nil {
					panic(err)
				}
				fmt.Println(test.fname)
				fmt.Printf("%v\n", p1.SampleType)
				fmt.Printf("%v\n", p2.SampleType)
				var scale []float64
				if len(p1.SampleType) == 4 {
					scale = []float64{-1, -1, 0, 0}
				} else {
					scale = []float64{-1, -1}
				}
				err = p1.ScaleN(scale)
				if err != nil {
					panic(err)
				}
				_, err = profile.Merge([]*profile.Profile{p1, p2})
				if err != nil {
					panic(err)
				}

			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
}
