FROM ubuntu:20.04

WORKDIR /

RUN apt-get update \
  && apt-get install -y libssl-dev

EXPOSE 8080
USER nobody
ENTRYPOINT ["exit 0"]
