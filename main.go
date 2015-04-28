package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"github.com/pelletier/go-toml"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	router        = mux.NewRouter()
	configuration = new(Configuration)
	config        *toml.TomlTree
	configErr     error
)

func main() {
	addr := flag.String("addr", ":8080", "Address in form of ip:port to listen on")
	certFile := flag.String("cert", "", "X.509 certificate file")
	keyFile := flag.String("key", "", "X.509 key file")
	public := flag.String("public", "", "Public directory to serve at the / endpoint")
	useXFF := flag.Bool("use-x-forwarded-for", false, "Use the X-Forwarded-For header when available")
	silent := flag.Bool("silent", false, "Do not log requests to stderr")
	awsKey := flag.String("awsKey", "fill your key", "Aws key string for amazon aws services")
	awsSecret := flag.String("awsSecret", "fill your secret/pLJqCIX", "Aws secret string for amazon aws services")
	awsRegion := flag.String("awsRegion", "us-east-1", "Amazon aws services location")
	awsBasePath := flag.String("awsBasePath", "s3.amazonaws.com/", "S3 base path")
	awsBucketPath := flag.String("awsBucketPath", "put the bucket path", "S3 file path in bucket")
	awsEndPoint := flag.String("awsEndPoint", "https://s3.amazonaws.com", "S3 end point domain")
	bucketName := flag.String("bucketName", "moon-bi", "The base s3 bucket")
	fileContentType := flag.String("fileContentType", "application/octet-stream", "S3 base path")
	environment := flag.String("env", "dev", "define environment for deployment : dev,qa,staging,all")
	pprof := flag.String("pprof", "", "Address in form of ip:port to listen on for pprof")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	/* Set the configuration */
	path, err_wd := os.Getwd()
	if err_wd != nil {
		log.Fatal(err_wd)
	}

	conf_path := fmt.Sprintf("%s/config/%s.conf", path, os.Getenv("ENV"))
	config, configErr = toml.LoadFile(conf_path)
	if configErr != nil {
		// Set the default configuration
		configuration.AwsKey = *awsKey
		configuration.AwsSecret = *awsSecret
		configuration.AwsBasePath = *awsBasePath
		configuration.AwsRegion = *awsRegion
		configuration.AwsBucketPath = *awsBucketPath
		configuration.AwsEndPoint = *awsEndPoint
		configuration.FileContentType = *fileContentType
		configuration.Environment = *environment
		configuration.BucketName = *bucketName
	}

	encoders := map[string]http.Handler{
		"/csv/":  NewHandler(&CSVEncoder{UseCRLF: true}),
		"/xml/":  NewHandler(&XMLEncoder{Indent: true}),
		"/json/": NewHandler(&JSONEncoder{}),
	}

	mux := http.NewServeMux()
	for path, handler := range encoders {
		mux.Handle(path, handler)
	}

	if len(*public) > 0 {
		mux.Handle("/", http.FileServer(http.Dir(*public)))
	}

	handler := Cors(mux, "GET", "HEAD")

	if !*silent {
		log.Println("go2aws server starting on", *addr)
		handler = logHandler(handler)
	}

	if *useXFF {
		handler = ProxyHandler(handler)
	}

	if len(*pprof) > 0 {
		go func() {
			log.Fatal(http.ListenAndServe(*pprof, nil))
		}()
	}

	var listenErr error
	if len(*certFile) > 0 && len(*keyFile) > 0 {
		listenErr = http.ListenAndServeTLS(*addr, *certFile, *keyFile, handler)
	} else {
		listenErr = http.ListenAndServe(*addr, handler)
	}
	if listenErr != nil {
		log.Fatal(listenErr)
	}
}

// Contain all the configuration from the pre-defined config
// file.
type Configuration struct {
	FilesPath       string
	DestFilesPath   string
	AwsKey          string
	AwsSecret       string
	AwsRegion       string
	AwsBasePath     string
	AwsBucketPath   string
	AwsEndPoint     string
	FileContentType string
	BucketName      string
	Environment     string
}

// A Handler provides http handlers that can process requests and return
// data in multiple formats.
//
// Usage:
//
// 	handle := NewHandler(encoder)
// 	http.Handle("/json/", handle.JSON())
//
// Note that the url pattern must end with a trailing slash since the
// handler looks for IP addresses or hostnames as parameters, for
// example /json/fileName.
type Handler struct {
	enc Encoder
	s3  *s3.S3
}

// NewHandler creates and initializes a new Handler with the specify encode and s3 object.
func NewHandler(enc Encoder) *Handler {
	var s *s3.S3
	auth := aws.Auth{configuration.AwsKey, configuration.AwsSecret, ""}
	s = s3.New(auth, aws.Region{Name: configuration.AwsRegion, S3Endpoint: configuration.AwsEndPoint})
	return &Handler{enc, s}
}

// ServeHTTP implements the http.Handler interface.
func (f *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the requested file from s3.
	fileName := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	bucket := f.s3.Bucket(configuration.BucketName)
	file, errGet := Get(bucket, configuration.AwsBucketPath+fileName)
	if errGet != nil {
		http.Error(w, "Try again later.",
			http.StatusServiceUnavailable)
		return
	}

	// Pass the file to the requested encoder and return the response.
	err := f.enc.Encode(w, r, file, fileName)

	if err != nil {
		http.Error(w, "An unexpected error occurred.",
			http.StatusInternalServerError)
		return
	}
}

// responseWriter is an http.ResponseWriter that records the returned
// status and bytes written to the client.
type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

// Write implements the http.ResponseWriter interface.
func (f *responseWriter) Write(b []byte) (int, error) {
	n, err := f.ResponseWriter.Write(b)
	if err != nil {
		return 0, err
	}
	f.bytes += n
	return n, nil
}

// WriteHeader implements the http.ResponseWriter interface.
func (f *responseWriter) WriteHeader(code int) {
	f.status = code
	f.ResponseWriter.WriteHeader(code)
}

// Get the raltive paths for key in bucket/folder.
func relativePath(path string, filePath string) string {
	if path == "." {
		return strings.TrimPrefix(filePath, "/")
	} else {
		return strings.TrimPrefix(strings.TrimPrefix(filePath, path), "/")
	}
}

// logHandler logs http requests.
func logHandler(f http.Handler) http.Handler {
	empty := ""
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := responseWriter{w, http.StatusOK, 0}
		start := time.Now()
		f.ServeHTTP(&resp, r)
		elapsed := time.Since(start)
		extra := context.Get(r, "log")
		if extra != nil {
			defer context.Clear(r)
		} else {
			extra = empty
		}
		log.Printf("%q %d %q %q %s %q %db in %s %q",
			r.Proto,
			resp.status,
			r.Method,
			r.URL.Path,
			r.Header.Get("User-Agent"),
			resp.bytes,
			elapsed,
			extra,
		)
	})
}

// CORS is an http handler that checks for allowed request methods (verbs)
// and adds CORS headers to all http responses.
//
// See http://en.wikipedia.org/wiki/Cross-origin_resource_sharing for details.
func Cors(f http.Handler, allow ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Method",
			strings.Join(allow, ", ")+", OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			return
		}
		for _, method := range allow {
			if r.Method == method {
				f.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("Allow", strings.Join(allow, ", ")+", OPTIONS")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed)
	})
}

// serviceUnavailable writes an http error 501 to a client.
func serviceUnavailable(w http.ResponseWriter, r *http.Request, log string) {
	context.Set(r, "log", log)
	http.Error(w, "Try again later", http.StatusServiceUnavailable)
}

// ProxyHandler is a wrapper for other http handlers that sets the
// client IP address in request.RemoteAddr to the first value of a
// comma separated list of IPs from the X-Forwarded-For request
// header. It resets the original RemoteAddr back after running the
// designated handler f.
func ProxyHandler(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addr := r.Header.Get("X-Forwarded-For")
		if len(addr) > 0 {
			remoteAddr := r.RemoteAddr
			r.RemoteAddr = strings.SplitN(addr, ",", 2)[0]
			defer func() { r.RemoteAddr = remoteAddr }()
		}
		f.ServeHTTP(w, r)
	})
}
