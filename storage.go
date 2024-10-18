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
	Save(context.Context, string, []byte) error
	Load(context.Context, string) (io.ReadCloser, error)
}

type LocalStorage struct {
	DataDir string
}

func NewLocalStorage(datadir string) *LocalStorage {
	return &LocalStorage{
		DataDir: datadir,
	}
}

func (s *LocalStorage) Save(_ context.Context, name string, data []byte) error {
	f := filepath.Join(s.DataDir, name+".txt")
	log.Println("[debug] save to", f)
	return os.WriteFile(f, data, 0644)
}

func (s *LocalStorage) Load(_ context.Context, name string) (io.ReadCloser, error) {
	f := filepath.Join(s.DataDir, name+".txt")
	log.Println("[debug] load from", f)
	return os.Open(f)
}

type S3Storage struct {
	Bucket    string
	KeyPrefix string
	svc       *s3.Client
}

func NewS3Storage(ctx context.Context, c *S3Config) (*S3Storage, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load config, %v", err)
	}
	svc := s3.NewFromConfig(cfg)
	return &S3Storage{
		Bucket:    c.Bucket,
		KeyPrefix: c.KeyPrefix,
		svc:       svc,
	}, nil
}

func (s *S3Storage) Load(ctx context.Context, name string) (io.ReadCloser, error) {
	for _, name := range []string{s.objectName(name), name} {
		result, err := s.svc.GetObject(ctx,
			&s3.GetObjectInput{
				Bucket: aws.String(s.Bucket),
				Key:    aws.String(path.Join(s.KeyPrefix, name)),
			},
		)
		log.Printf("[debug] load from s3://%s", path.Join(s.Bucket, s.KeyPrefix, name))
		if err == nil {
			log.Printf("[debug] result %v", result)
			return result.Body, nil
		}
	}
	return nil, fmt.Errorf("not found %s and %s", name, s.objectName(name))
}

func (s *S3Storage) Save(ctx context.Context, name string, b []byte) error {
	name = s.objectName(name)
	input := &s3.PutObjectInput{
		Body:        bytes.NewReader(b),
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(path.Join(s.KeyPrefix, name)),
		ContentType: aws.String("text/plain"),
	}
	log.Printf("[debug] save to s3://%s", path.Join(s.Bucket, s.KeyPrefix, name))
	_, err := s.svc.PutObject(ctx, input)
	return err
}

func (s *S3Storage) objectName(name string) string {
	if len(name) > 5 {
		return path.Join(name[0:2], name[2:4], name[4:])
	} else {
		return name
	}
}
