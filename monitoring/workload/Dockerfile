ARG ARCH
FROM $ARCH/ubuntu:latest

RUN apt update && \
    apt install -y gcc bc cgroup-tools python3-pip

# AMS object storage
RUN pip3 install awscli awscli-plugin-endpoint

RUN aws configure set plugins.endpoint awscli_plugin_endpoint && \
    aws configure set aws_access_key_id SCW9PYDH2VX3EWYZ0JD7 && \
    aws configure set aws_secret_access_key d13a460c-dc4c-43e5-abf9-b0bfb24e4b40

COPY aws_config ~/.aws/config
# COPY aws_credentials ~/.aws/credentials
# cpulimit cgroup-tools

COPY gen_stress.sh /
COPY trace.data /
COPY stress_worker.c /

WORKDIR /
RUN gcc stress_worker.c -o stress_worker

CMD ["bash", "gen_stress.sh"]
