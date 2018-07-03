FROM ubuntu:16.04
RUN apt-get update && apt-get install -y ca-certificates curl && rm -rf /var/cache/apt
COPY ./ahabd /usr/bin/ahabd
ENTRYPOINT ["/usr/bin/ahabd"]
