FROM amazon/aws-cli

RUN yum install -y jq && yum clean all
