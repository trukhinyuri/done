# Create final image
FROM ubuntu:latest

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl wget unzip zip make

RUN curl -s "https://get.sdkman.io" | bash

RUN bash -c ". $HOME/.sdkman/bin/sdkman-init.sh && sdk install java 21.0.1-oracle"
RUN bash -c ". $HOME/.sdkman/bin/sdkman-init.sh && sdk install kotlin 1.9.21"
RUN bash -c ". $HOME/.sdkman/bin/sdkman-init.sh && sdk install gradle 8.5"
