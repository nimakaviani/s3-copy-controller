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

package controllers

import (
	ctrlapi "dev.nimak.link/s3-copy-controller/controllers/api"
	awshelper "dev.nimak.link/s3-copy-controller/controllers/aws"
)

type storeManager struct{}

func NewStoreManager() ctrlapi.StoreManager {
	return &storeManager{}
}

func (s storeManager) Get(cfg ctrlapi.ConfigData) ctrlapi.ObjectStore {
	return awshelper.NewS3ObjectStore(cfg)
}
