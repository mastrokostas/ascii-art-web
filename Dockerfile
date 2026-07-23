FROM golang:1.25.0

LABEL org.opencontainers.image.title="Ascii Art Web Dockerize" \
      org.opencontainers.image.description="Runs a dockerized version of the Ascii Art Web project"

WORKDIR /app

COPY . /app

RUN go build -o ascii-art-web-dockerize .

RUN useradd -u 1001 ascii-art-web

USER ascii-art-web

EXPOSE 8080

CMD [ "./ascii-art-web-dockerize" ]