FROM ubuntu:18.04
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/cache/apt
ADD https://storage.googleapis.com/kubernetes-release/release/v1.10.5/bin/linux/amd64/kubectl /usr/bin/kubectl
RUN chmod 0755 /usr/bin/kubectl
COPY ./ahabd /usr/bin/ahabd
ENTRYPOINT ["/usr/bin/ahabd"]
