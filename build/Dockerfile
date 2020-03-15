FROM centos
LABEL maintainers="Kubernetes Local Volume"
LABEL description="Kubernetes Local Volume CSI Plugin"

# install lvm
RUN yum update -y && yum install lvm2 -y && yum install file -y && yum install e4fsprogs -y

# add nsenter
COPY nsenter /nsenter

# add plugin
COPY _output/local.volume.csi.kubernetes.io /bin/local.volume.csi.kubernetes.io
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /bin/local.volume.csi.kubernetes.io && chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
