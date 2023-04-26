ARG ARCH=amd64
FROM --platform=linux/${ARCH} gcr.io/distroless/static-debian11

ADD etcd-defrag /usr/local/bin/

CMD ["/usr/local/bin/etcd-defrag"]
