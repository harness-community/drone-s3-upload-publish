# drone-s3-upload-publish

Drone plugin to upload file/directories to AWS S3 Bucket and display the bucket url under 'Executions > Artifacts' tab.

## Build

Build the binary with the following commands:

```bash
go build
```

## Docker

Build the Docker image with the following commands:

```
./scripts/build.sh
docker buildx build -t DOCKER_ORG/drone-s3-upload-publish --platform linux/amd64 .
```

Please note incorrectly building the image for the correct x64 linux and with
CGO disabled will result in an error when running the Docker image:

```
docker: Error response from daemon: Container command
'/bin/drone-s3-upload-publish' not found or does not exist..
```

## Usage

```bash
docker run --rm \
  -e PLUGIN_AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
  -e PLUGIN_AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
  -e PLUGIN_AWS_DEFAULT_REGION=ap-southeast-2 \
  -e PLUGIN_AWS_BUCKET=bucket-name \
  -e PLUGIN_SOURCE=OBJECT_PATH \
  -e PLUGIN_ARTIFACT_FILE=url.txt \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  harnesscommunity/drone-s3-upload-publish
```

In Harness CI,
```yaml
              - step:
                  type: Plugin
                  name: s3-upload-publish
                  identifier: custom_plugin
                  spec:
                    connectorRef: account.harnessImage
                    image: harnesscommunity/drone-s3-upload-publish
                    settings:
                      aws_access_key_id: <+pipeline.variables.AWS_ACCESS>
                      aws_secret_access_key: <+pipeline.variables.AWS_SECRET>
                      aws_default_region: ap-southeast-2
                      aws_bucket: bucket-name
                      artifact_file: url.txt
                      source: OBJECT_PATH
                    imagePullPolicy: IfNotPresent
```
## Include files with glob pattern Usage
```bash
docker run --rm --network host \
  -e   PLUGIN_AWS_ACCESS_KEY_ID=$HNS_AWS_ACCESS_KEY_ID \
  -e   PLUGIN_AWS_SECRET_ACCESS_KEY=$HNS_AWS_SECRET_ACCESS_KEY \
  -e   PLUGIN_AWS_DEFAULT_REGION=ap-south-1 \
  -e   PLUGIN_AWS_BUCKET=hns-test-bucket \
  -e   PLUGIN_SOURCE=./test \
  -e   PLUGIN_ARTIFACT_FILE=url.txt \
  -e   PLUGIN_TARGET=test20241111503 \
  -e   PLUGIN_GLOB='**/*.html, **/*.css' \
  harnesscommunity/drone-s3-upload-publish
  ```
In Harness CI, for include files with glob pattern Usage
```yaml
              - step:
                  type: Plugin
                  name: html_publish_step
                  identifier: html_publish_step
                  spec:
                    connectorRef: Docker_Hub_Anonymous
                    image: senthilhns/html_publish_01:latest
                    settings:
                      artifact_file: artifact.txt
                      aws_access_key_id:  <+pipeline.variables.AWS_ACCESS>
                      aws_secret_access_key:  <+pipeline.variables.AWS_SECRET>
                      aws_bucket: hns-test-bucket
                      aws_default_region: ap-southeast-2
                      glob: "**/*.html, **/*.css"
                      source: ./test
                      target: test2014am23_css_html_only
```

To get the list of supported arguments:
```bash
go build

./drone-s3-upload-publish --help
```
```
NAME:
   drone-s3-upload-publish - Drone plugin to upload file/directories to AWS S3 Bucket and display the bucket url under 'Executions > Artifacts' tab
...
...
GLOBAL OPTIONS:
   --aws-access-key value      AWS Access Key ID [$PLUGIN_AWS_ACCESS_KEY_ID]
   --aws-secret-key value      AWS Secret Access Key [$PLUGIN_AWS_SECRET_ACCESS_KEY]
   --aws-default-region value  AWS Default Region [$PLUGIN_AWS_DEFAULT_REGION]
   --aws-bucket value          AWS S3 Bucket [$PLUGIN_AWS_BUCKET]
   --source value              Source [$PLUGIN_SOURCE]
   --target-path value         target [$PLUGIN_TARGET]
   --artifact-file value       Artifact file [$PLUGIN_ARTIFACT_FILE]
   --glob value                Include file patterns int ant style glob style [$PLUGIN_INCLUDE]
   --help, -h                  show help
   --version, -v               print the version
...
...
```