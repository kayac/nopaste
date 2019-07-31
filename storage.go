package nopaste

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Storage interface {
	Save(string, []byte) error
	Load(string) (io.ReadCloser, error)
}

type LocalStorage struct {
	DataDir string
}

func NewLocalStorage(datadir string) *LocalStorage {
	return &LocalStorage{
		DataDir: datadir,
	}
}

func (s *LocalStorage) Save(name string, data []byte) error {
	f := filepath.Join(s.DataDir, name+".txt")
	log.Println("[debug] save to", f)
	return ioutil.WriteFile(f, data, 0644)
}

func (s *LocalStorage) Load(name string) (io.ReadCloser, error) {
	f := filepath.Join(s.DataDir, name+".txt")
	log.Println("[debug] load from", f)
	return os.Open(f)
}

type S3Storage struct {
	Bucket    string
	KeyPrefix string
	svc       *s3.S3
}

func NewS3Storage(c *S3Config) *S3Storage {
	sess := session.Must(session.NewSession())
	svc := s3.New(sess)
	return &S3Storage{
		Bucket:    c.Bucket,
		KeyPrefix: c.KeyPrefix,
		svc:       svc,
	}
}

func (s *S3Storage) Load(name string) (io.ReadCloser, error) {
	result, err := s.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(path.Join(s.KeyPrefix, name)),
	})
	log.Printf("[debug] load from s3://%s", path.Join(s.Bucket, s.KeyPrefix, name))
	if err != nil {
		return nil, err
	}
	log.Println("[debug] result", result.GoString())
	return result.Body, nil
}

func (s *S3Storage) Save(name string, b []byte) error {
	input := &s3.PutObjectInput{
		Body:        aws.ReadSeekCloser(bytes.NewReader(b)),
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(path.Join(s.KeyPrefix, name)),
		ContentType: aws.String("text/plain"),
	}
	log.Printf("[debug] save to s3://%s", path.Join(s.Bucket, s.KeyPrefix, name))
	_, err := s.svc.PutObject(input)
	return err
}
