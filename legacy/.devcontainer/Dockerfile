FROM ubuntu:20.04

# Set environment variables to non-interactive (this prevents some prompts)
ENV DEBIAN_FRONTEND=noninteractive

# Install Docker
RUN apt update

RUN apt install curl -y

RUN curl -sSL https://get.docker.com/ | sh

RUN apt-get install -y curl && \
    apt-get install -y sudo && \
    apt-get clean

#RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
#    chmod +x kubectl && \
#    mv kubectl /usr/local/bin/
#
#RUN curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64 && \
#    chmod +x ./kind && \
#    mv kind /usr/local/bin/


ENTRYPOINT service docker start && kind create cluster && /bin/bash
