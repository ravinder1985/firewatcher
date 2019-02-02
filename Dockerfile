FROM        centos:7
MAINTAINER  Harvinder Singh <ravinderkarnana@gmail.com>

ARG version=1.1.4
ARG application=firewatcher_linux_amd64

RUN yum clean all
RUN yum -y upgrade
RUN yum update -y && yum install -y \
                        curl \
                        unzip \
                        tar \
                        wget \
                        git \
                        sed
RUN yum clean all && rm -rf /var/cache/yum
RUN wget https://github.com/ravinder1985/firewatcher/releases/download/${application}_${version}/${application}_${version}.tar -O /tmp/${application}_${version}.tar \
	&& cd /tmp && tar -xvf ${application}_${version}.tar \
	&& mkdir /opt/firewatcher && mkdir /opt/firewatcher/scripts && cp -R /tmp/${application} /opt/firewatcher/${application}
ADD config.json /opt/firewatcher/config.json
ADD scripts/* /opt/firewatcher/scripts/

EXPOSE     8080
WORKDIR    /opt/firewatcher
ENTRYPOINT [ "/opt/firewatcher/firewatcher_linux_amd64" ]
CMD        [ "-config=/opt/firewatcher/config.json", \
             "-storage.path=/opt/firewatcher"

