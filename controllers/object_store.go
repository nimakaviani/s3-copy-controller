package controllers

import (
	"context"

	cloudobject "dev.nimak.link/s3-copy-controller/api/v1alpha1"
)

type ObjectStore interface {
	Store(context.Context, []byte, cloudobject.ObjectTarget) error
}
