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
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/go-ini/ini"
	"github.com/pkg/errors"
)

const (
	defaultProfile = "default"
)

// StringValue converts the supplied string pointer to a string, returning the
// empty string if the pointer is nil.
func StringValue(v *string) string {
	return aws.ToString(v)
}

// CredentialsIDSecret retrieves AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY from the data which contains
// aws credentials under given profile
// Example:
// [default]
// aws_access_key_id = <YOUR_ACCESS_KEY_ID>
// aws_secret_access_key = <YOUR_SECRET_ACCESS_KEY>
func credentialsIDSecret(data []byte, profile string) (aws.Credentials, error) {
	config, err := ini.InsensitiveLoad(data)
	if err != nil {
		return aws.Credentials{}, errors.Wrap(err, "cannot parse credentials secret")
	}

	iniProfile, err := config.GetSection(profile)
	if err != nil {
		return aws.Credentials{}, errors.Wrap(err, fmt.Sprintf("cannot get %s profile in credentials secret", profile))
	}

	accessKeyID := iniProfile.Key("aws_access_key_id")
	secretAccessKey := iniProfile.Key("aws_secret_access_key")
	sessionToken := iniProfile.Key("aws_session_token")

	// NOTE(muvaf): Key function implementation never returns nil but still its
	// type is pointer so we check to make sure its next versions doesn't break
	// that implicit contract.
	if accessKeyID == nil || secretAccessKey == nil || sessionToken == nil {
		return aws.Credentials{}, errors.New("returned key can be empty but cannot be nil")
	}

	return aws.Credentials{
		AccessKeyID:     accessKeyID.Value(),
		SecretAccessKey: secretAccessKey.Value(),
		SessionToken:    sessionToken.Value(),
	}, nil
}

// UseProviderSecret - AWS configuration which can be used to issue requests against AWS API
func useProviderSecret(ctx context.Context, data []byte, region, profile string) (*aws.Config, error) {
	creds, err := credentialsIDSecret(data, profile)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse credentials secret")
	}

	config, err := config.LoadDefaultConfig(ctx, config.WithRegion(region), config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
		Value: creds,
	}))
	return &config, err
}
