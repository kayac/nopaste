package nopaste

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
	return os.WriteFile(f, data, 0644)
}

func (s *LocalStorage) Load(name string) (io.ReadCloser, error) {
	f := filepath.Join(s.DataDir, name+".txt")
	log.Println("[debug] load from", f)
	return os.Open(f)
}

type S3Storage struct {
	Bucket    string
	KeyPrefix string
	svc       *s3.Client
}

func NewS3Storage(c *S3Config) *S3Storage {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	svc := s3.NewFromConfig(cfg)
	return &S3Storage{
		Bucket:    c.Bucket,
		KeyPrefix: c.KeyPrefix,
		svc:       svc,
	}
}

func (s *S3Storage) Load(name string) (io.ReadCloser, error) {
	for _, name := range []string{s.objectName(name), name} {
		result, err := s.svc.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(s.Bucket),
			Key:    aws.String(path.Join(s.KeyPrefix, name)),
		})
		log.Printf("[debug] load from s3://%s", path.Join(s.Bucket, s.KeyPrefix, name))
		if err == nil {
			log.Printf("[debug] result %v", result)
			return result.Body, nil
		}
	}
	return nil, fmt.Errorf("not found %s and %s", name, s.objectName(name))
}

func (s *S3Storage) Save(name string, b []byte) error {
	name = s.objectName(name)
	input := &s3.PutObjectInput{
		Body:        bytes.NewReader(b),
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(path.Join(s.KeyPrefix, name)),
		ContentType: aws.String("text/plain"),
	}
	log.Printf("[debug] save to s3://%s", path.Join(s.Bucket, s.KeyPrefix, name))
	_, err := s.svc.PutObject(context.TODO(), input)
	return err
}

func (s *S3Storage) objectName(name string) string {
	if len(name) > 5 {
		return path.Join(name[0:2], name[2:4], name[4:])
	} else {
		return name
	}
}
