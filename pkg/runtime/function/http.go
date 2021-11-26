package function

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type httpHeader http.Header

func (h *httpHeader) String() string {
	sb := strings.Builder{}
	for k, v := range *h {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(strings.Join(v, "; "))
	}
	return sb.String()
}

func (h *httpHeader) Set(s string) error {
	ss := strings.Split(s, ":")
	if *h == nil {
		*h = make(httpHeader)
	}
	http.Header(*h).Add(ss[0], ss[1])
	return nil
}

type httpOption struct {
	header  httpHeader
	data    string
	dataSet bool
	args    []string
	*options
}

func (ho *httpOption) reset() {
	ho.header = nil
	ho.args = nil
	ho.data = ""
	ho.dataSet = false
}

func (ho *httpOption) copy() interface{} {
	return &httpOption{
		header:  ho.header,
		data:    ho.data,
		dataSet: ho.dataSet,
		args:    ho.options.args,
	}
}

func (ho *httpOption) flagHeader(fs *flag.FlagSet) {
	fs.Var(&ho.header, "h", "custom http header to be include or override existing header in the request")
}

func (ho *httpOption) validate(gf *GeneralFunction) error {
	if len(ho.args) != 1 {
		return fmt.Errorf("%s function required a single url argument", gf.name)
	}
	return nil
}

func (ho *httpOption) Parse(fs *flag.FlagSet, args []string) (i interface{}, err error) {
	if i, err = ho.options.Parse(fs, args); err == nil {
		opts := i.(*httpOption)
		opts.dataSet = false
		fs.Visit(func(f *flag.Flag) {
			if f.Name == "d" {
				opts.dataSet = true
			}
		})
	}
	return
}

func newHttpOptions(fs *flag.FlagSet, wantBody bool) *httpOption {
	opts := &httpOption{}
	opts.options = &options{opts: opts}
	opts.flagHeader(fs)
	if wantBody {
		flagString(fs, &opts.data, "d", string([]byte{0}), "post data as a string or a path to file if it start with @")
	}
	return opts
}

type readerCloser struct {
	*bytes.Reader
}

func (rc *readerCloser) Close() error { return nil }

func detectContentType(r io.ReadSeekCloser) string {
	defer r.Seek(0, 0)
	buf := [512]byte{}
	return http.DetectContentType(buf[:])
}

// for override testing purpose only
var returnFunc = func(resp *http.Response, canHasResponseBody bool) interface{} {
	if canHasResponseBody {
		return resp.Body
	}
	return nil
}

func httpRequest(gf *GeneralFunction, i interface{}, method string) (result interface{}, err error) {
	opts := i.(*httpOption)
	if err = opts.validate(gf); err != nil {
		return nil, err
	}
	var resp *http.Response
	var req *http.Request
	var body io.ReadSeekCloser
	var canResponseHasBody = false
	switch method {
	case http.MethodOptions, http.MethodGet:
		canResponseHasBody = true
	case http.MethodDelete:
		canResponseHasBody = true
		if opts.data == "" {
			goto createreq
		}
		fallthrough
	case http.MethodPost, http.MethodPatch:
		canResponseHasBody = true
		fallthrough
	case http.MethodPut:
		if opts.dataSet {
			if strings.HasPrefix(opts.data, "@") {
				body, err = os.Open(opts.data)
				if err != nil {
					return nil, err
				}
			} else {
				body = &readerCloser{Reader: bytes.NewReader([]byte(opts.data))}
			}
		}
	}
createreq:
	if req, err = http.NewRequest(method, opts.args[0], body); err != nil {
		return nil, err
	}
	// set header if available
	if opts.header != nil {
		req.Header = http.Header(opts.header)
	}
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", detectContentType(body))
	}
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return nil, err
	}
	switch {
	case resp.StatusCode == http.StatusOK:
		return returnFunc(resp, canResponseHasBody), nil
	case resp.StatusCode < 300:
		return nil, nil
	default:
		return nil, fmt.Errorf("server report %d on get %s", resp.StatusCode, opts.args[0])
	}
}

func init() {
	registerFunction(&GeneralFunction{
		name:     "get", // @get -h k:v -h k:v http://www.example.com
		flagInit: func(fs *flag.FlagSet) Option { return newHttpOptions(fs, false) },
		handler: func(gf *GeneralFunction, i interface{}) (result interface{}, err error) {
			return httpRequest(gf, i, http.MethodGet)
		},
	})

	registerFunction(&GeneralFunction{
		name:     "head",
		flagInit: func(fs *flag.FlagSet) Option { return newHttpOptions(fs, false) },
		handler: func(gf *GeneralFunction, i interface{}) (result interface{}, err error) {
			return httpRequest(gf, i, http.MethodHead)
		},
	})

	registerFunction(&GeneralFunction{
		name:     "options",
		flagInit: func(fs *flag.FlagSet) Option { return newHttpOptions(fs, false) },
		handler: func(gf *GeneralFunction, i interface{}) (result interface{}, err error) {
			return httpRequest(gf, i, http.MethodOptions)
		},
	})

	registerFunction(&GeneralFunction{
		name:     "post", // @post -h k:v -d {@file|string} http://www.example.com
		flagInit: func(fs *flag.FlagSet) Option { return newHttpOptions(fs, true) },
		handler: func(gf *GeneralFunction, i interface{}) (result interface{}, err error) {
			return httpRequest(gf, i, http.MethodPost)
		},
	})

	registerFunction(&GeneralFunction{
		name:     "patch",
		flagInit: func(fs *flag.FlagSet) Option { return newHttpOptions(fs, true) },
		handler: func(gf *GeneralFunction, i interface{}) (result interface{}, err error) {
			return httpRequest(gf, i, http.MethodPatch)
		},
	})

	registerFunction(&GeneralFunction{
		name:     "put",
		flagInit: func(fs *flag.FlagSet) Option { return newHttpOptions(fs, true) },
		handler: func(gf *GeneralFunction, i interface{}) (result interface{}, err error) {
			return httpRequest(gf, i, http.MethodPut)
		},
	})

	registerFunction(&GeneralFunction{
		name:     "delete",
		flagInit: func(fs *flag.FlagSet) Option { return newHttpOptions(fs, true) },
		handler: func(gf *GeneralFunction, i interface{}) (result interface{}, err error) {
			return httpRequest(gf, i, http.MethodDelete)
		},
	})
}
