FROM amazon/aws-cli:amd64

LABEL maintainer="OP (ompragash) <ompragash@proton.me>"

ADD release/linux/amd64/drone-s3-upload-publish /bin/

ENTRYPOINT ["/usr/bin/bash", "-l", "-c", "/bin/drone-s3-upload-publish"]
