package main

import (
	"github.com/mitchellh/goamz/s3"
	"strings"
)

// Helper class to work with s3 objects
type S3Url struct {
	Url string
}

// Get the bucket name from the retured keys
func (r *S3Url) Bucket() string {
	return r.keys()[0]
}

// Get the full key name from the retured keys
func (r *S3Url) Key() string {
	return strings.Join(r.keys()[1:len(r.keys())], "/")
}

// Get the full path for the key
func (r *S3Url) Path() string {
	return r.Key()
}

// Check if the current key is valid, which mean it contain the prefix s3://
func (r *S3Url) Valid() bool {
	return strings.HasPrefix(r.Url, "s3://")
}

// Get list of keys
func (r *S3Url) keys() []string {
	trimmed_string := strings.TrimLeft(r.Url, "s3://")
	return strings.Split(trimmed_string, "/")
}

// Get object(key) from bucket & path
func Get(bucket *s3.Bucket, path string) (data []byte, er error) {
	data, err := bucket.Get(path)
	if err != nil {
		panic(err.Error())
	}
	return data, err
}
