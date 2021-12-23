package aws

import (
	"bytes"
	"context"

	cloudobject "dev.nimak.link/s3-copy-controller/api/v1alpha1"
	ctrlapi "dev.nimak.link/s3-copy-controller/controllers/api"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3ObjectStore struct {
	cfg    *aws.Config
	client *s3.Client
}

func NewS3ObjectStore(cfg *aws.Config) ctrlapi.ObjectStore {
	return &s3ObjectStore{
		cfg:    cfg,
		client: s3.NewFromConfig(*cfg),
	}
}

func (s *s3ObjectStore) Store(ctx context.Context, content []byte, target cloudobject.ObjectTarget) error {
	input := &s3.PutObjectInput{
		Bucket: &target.Bucket,
		Key:    &target.Key,
		Body:   bytes.NewReader(content),
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

func (s *s3ObjectStore) Delete(ctx context.Context, target cloudobject.ObjectTarget) error {
	input := &s3.DeleteObjectInput{
		Bucket: &target.Bucket,
		Key:    &target.Key,
	}

	if _, err := s.client.DeleteObject(ctx, input); err != nil {
		return err
	}

	return nil
}
