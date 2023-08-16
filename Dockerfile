FROM golang:latest AS buildStage
RUN apt-get update
COPY . /app
WORKDIR /app
RUN go build -o done .

FROM ubuntu:latest
COPY --from=buildStage /app/done /app/
COPY --from=buildStage /app/frontend/ /app/frontend/
WORKDIR /app
CMD ./done
