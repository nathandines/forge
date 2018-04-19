# Using alpine as this contains the appropriate CA certificates which are
# needed ongoing to authenticate the AWS API servers
FROM alpine:latest

COPY bin/forge /usr/bin/forge

RUN mkdir /workdir
WORKDIR /workdir

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
