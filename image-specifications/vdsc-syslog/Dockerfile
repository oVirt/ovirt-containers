FROM ovirt/base:master

MAINTAINER "Yaniv Bronhaim" <ybronhei@redhat.com>

RUN yum -y install rsyslog && yum -y clean all

EXPOSE 514 514/tcp

COPY root /

ENTRYPOINT /entrypoint.sh
