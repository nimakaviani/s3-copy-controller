package controllers

import (
	"bytes"
	"context"

	cloudobject "dev.nimak.link/s3-copy-controller/api/v1alpha1"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3ObjectStore struct {
	cfg *aws.Config
}

func NewS3ObjectStore(cfg *aws.Config) ObjectStore {
	return &s3ObjectStore{
		cfg: cfg,
	}
}

func (s *s3ObjectStore) Store(ctx context.Context, content []byte, target cloudobject.ObjectTarget) error {
	client := s3.NewFromConfig(*s.cfg)

	input := &s3.PutObjectInput{
		Bucket: &target.Bucket,
		Key:    &target.Key,
		Body:   bytes.NewReader(content),
	}

	_, err := client.PutObject(ctx, input)
	if err != nil {
		return err
	}

	return nil
}
