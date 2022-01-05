# S3 Copy Controller

_S3CopyController_ is a data plane Kubernetes controller that allows data from custom
Kubernetes objects or `ConfigMaps` to be saved to a cloud Object Store (for
now, AWS S3 only).

The controller is built using [KubeBuilder](https://github.com/kubernetes-sigs/kubebuilder) and is
in an experimental state.

## Installation

### From Source

You should be able to use KubeBuilder's internal scripts to deploy directly
against your Kubernetes cluster using the following command:

```
git clone https://github.com/nimakaviani/s3-copy-controller.git

cd s3-copy-controller

make deploy
```

Run the following command to verify the deployment

```shell script
kubectl api-resources | grep objects
```

Output should like below 

    NAME     SHORTNAMES                 APIGROUP                       NAMESPACED   KIND
    objects s3.aws.dev.nimak.link       true Object

or you can alternatively run Kustomize `build`, and deploy resources:

```
kustomize build config/default | kube apply -f -
```

## Usage

With the controller running on your cluster, you need to first provide AWS login
credentials in the form of a secret to your cluster:

### Step1: Create Base64 Encoded AWS Credentials

Create a text file with the actual AWS credentials

`creds.txt`
```sh
AWS_ACCESS_KEY_ID=ASIAUHGDERSWSWPYVM2DDUMMYDUMMY
AWS_SECRET_ACCESS_KEY=ZBo0Q0lDUMMYNyL5DUMMYDUMMY
AWS_SESSION_TOKEN=RYRHHVGVNNNDUMMYDUMMYDUMMYDUMMY
```

And run the following command from Mac command line

```sh
openssl base64 -in creds.txt -out base64-creds.txt
```

Copy the output as a string from `base64-creds.txt` and use it in Secret object under `aws.creds`

### Step2: Create `Secret` Object 

Create Secret object using `kubectl apply` with the following content

```yaml
apiVersion: v1
data:
  aws.creds: <Enter BASE64 Encoded AWS Credentials>
kind: Secret
metadata:
  name: aws-account-creds
  namespace: default
type: Opaque
```

### Step3: Deploy Sample `Object` 

Create sample Object for S3 Copy using `kubectl apply`

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
    bucket: <Enter S3 Bucket Name>
    key: <S3 Prefix>/<filename>.txt
  credentials:
    source: Secret
    secretRef:
      namespace: default
      name: aws-account-creds
      key: aws.creds
```

Submitting the following resource to your Kubernetes cluster should result in an
object getting created under `s3://<YourBucketName>/<S3 Prefix>/<filename>.txt`, with its
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
