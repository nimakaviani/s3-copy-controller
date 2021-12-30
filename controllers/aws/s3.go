/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package aws

import (
	"bytes"
	"context"

	cloudobject "dev.nimak.link/s3-copy-controller/api/v1alpha1"
	ctrlapi "dev.nimak.link/s3-copy-controller/controllers/api"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3ObjectStore struct {
	config ctrlapi.ConfigData
}

func NewS3ObjectStore(config ctrlapi.ConfigData) ctrlapi.ObjectStore {
	return &s3ObjectStore{
		config: config,
	}
}

func (s *s3ObjectStore) Store(ctx context.Context, content []byte, target cloudobject.ObjectTarget) error {
	cfg, err := useProviderSecret(ctx, s.config.Secret, s.config.Region, defaultProfile)
	if err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Bucket: &target.Bucket,
		Key:    &target.Key,
		Body:   bytes.NewReader(content),
	}

	client := s3.NewFromConfig(*cfg)
	_, err = ctrlapi.PutItem(ctx, client, input)
	if err != nil {
		return err
	}

	return nil
}

func (s *s3ObjectStore) Delete(ctx context.Context, target cloudobject.ObjectTarget) error {
	cfg, err := useProviderSecret(ctx, s.config.Secret, s.config.Region, defaultProfile)
	if err != nil {
		return err
	}

	input := &s3.DeleteObjectInput{
		Bucket: &target.Bucket,
		Key:    &target.Key,
	}

	client := s3.NewFromConfig(*cfg)
	if _, err = ctrlapi.DeleteItem(ctx, client, input); err != nil {
		return err
	}

	return nil
}
