ARG ARCH
FROM $ARCH/ubuntu:latest

RUN apt update && \
    apt install -y gcc bc cgroup-tools python3-pip

# AMS object storage
RUN pip3 install awscli awscli-plugin-endpoint && \
    aws configure set plugins.endpoint awscli_plugin_endpoint
COPY aws_config ~/.aws/config
COPY aws_credentials ~/.aws/credentials
# cpulimit cgroup-tools

COPY gen_stress.sh /
COPY trace.data /
COPY stress_worker.c /

WORKDIR /
RUN gcc stress_worker.c -o stress_worker

CMD ["bash", "gen_stress.sh"]
