ARG ARCH
FROM $ARCH/ubuntu:latest

RUN apt update && \
    apt install -y gcc python3-pip

# AMS object storage
RUN pip3 install awscli awscli-plugin-endpoint --break-system-packages

RUN aws configure set plugins.endpoint awscli_plugin_endpoint
    # aws configure set aws_access_key_id SCW9PYDH2VX3EWYZ0JD7 && \
    # aws configure set aws_secret_access_key d13a460c-dc4c-43e5-abf9-b0bfb24e4b40

COPY aws_config /root/.aws/config
COPY aws_credentials /root/.aws/credentials
# cpulimit cgroup-tools

# COPY gen_stress.sh /
COPY trace.data /
COPY stress_generator.c /

WORKDIR /
RUN gcc stress_generator.c -o stress_generator

# CMD ["bash", "gen_stress.sh"]
