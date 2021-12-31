# S3 Copy Controller

_S3CopyController_ is a data plane controller that allows data from custom
Kubernetes objects or `ConfigMaps` to be saved to a cloud Object Store (for
now, AWS S3 only).

The controller is built using [KubeBuilder](https://github.com/kubernetes-sigs/kubebuilder) and is
in an experimental state.

## Installation

### From Source

You should be able to use KubeBuilder's internal scripts to deploy directly
against your Kubernetes cluster using the following command:

```
make deploy
```

## Usage

With the controller running on your cluster, you need to first provide AWS login
credentials in the form of a secret to your cluster:

```yaml
apiVersion: v1
data:
  aws.creds: [BASE64 Encoded Credentials]
kind: Secret
metadata:
  name: aws-account-creds
  namespace: default
type: Opaque
```

With the secrets deployed, a sample `Object` resource looks like the following:

```yaml
apiVersion: s3.aws.dev.nimak.link/v1alpha1
kind: Object
metadata:
  name: sample
  namespace: default
spec:
  deletionPolicy: Delete # Delete / Retain are the options
  source:
    data: |
      something something
      and more ...
  target:
    region: us-west-2
    bucket: some-bucket
    key: scripts/data.txt
  credentials:
    source: Secret
    secretRef:
      namespace: default
      name: aws-account-creds
      key: aws.creds
```

Submitting the following resource to your Kubenretes cluster should result in an
object getting created under `s3://some-bucket/scripts/data.txt`, with its
content coming from `Spec.Source.Data` in your `sample` Object above.

## Development

The development process follows general practices for KubeBuilder.

- To generate CRDs and install them to the cluster, modify [the source
  object](/api/v1alpha1/object_types.go) and run:

```
make manifests && make install
```

- To build the code:

```
make build && make run
```

- To Run tests:
```
make tests
```
