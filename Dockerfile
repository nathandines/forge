# Using alpine as this contains the appropriate CA certificates which are
# needed ongoing to authenticate the AWS API servers
FROM alpine:latest

COPY bin/forge /usr/bin/forge

# Fix for musl vs glibc library availability on Alpine Linux
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

RUN mkdir /workdir
WORKDIR /workdir

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
