FROM alpine:3.8

ENV OPERATOR=/usr/local/bin/s3-operator \
    USER_UID=1001 \
    USER_NAME=s3-operator

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

# install operator binary
COPY build/_output/bin/s3-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
