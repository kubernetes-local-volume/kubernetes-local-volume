FROM ubuntu:18.10
LABEL maintainers="Kubernetes Local Volume"
LABEL description="Kubernetes Local Volume CSI Plugin"

RUN apt update && apt upgrade -y && apt install -y ca-certificates file tzdata
COPY nsenter /nsenter

COPY _output/local.volume.csi.kubernetes.io /bin/local.volume.csi.kubernetes.io
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /bin/local.volume.csi.kubernetes.io && chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
