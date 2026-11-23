package s3store

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
)

type Store struct {
	Client *s3.Client
	Bucket string
}

func New(cfg aws.Config, bucket string) *Store {
	return &Store{Client: s3.NewFromConfig(cfg), Bucket: bucket}
}
func (s *Store) Put(ctx context.Context, key string, body io.Reader, size int64) error {
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(s.Bucket), Key: aws.String(key), Body: body, ContentLength: &size, ContentType: aws.String("application/vnd.apache.parquet")})
	return err
}
