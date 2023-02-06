FROM golang

WORKDIR /tmp/build
RUN mkdir -p /tmp/build
COPY entrypoint.sh /tmp/build/entrypoint.sh
RUN chmod +x /tmp/build/entrypoint.sh
ENTRYPOINT /tmp/build/entrypoint.sh
