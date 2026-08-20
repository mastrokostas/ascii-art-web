FROM golang:1.26.0

LABEL org.opencontainers.image.title="Ascii Art Web" \
      org.opencontainers.image.description="Runs a dockerized version of the Ascii Art Web project"

WORKDIR /app

COPY . /app

RUN go build -o ascii-art-web .

RUN useradd -u 1001 ascii-art-web

# create the logs folder in the project and change owner to 1001
# MkdirAll succeeds trivially, and the bind mount lands on top of it in production
# removing this fucks the smoke testing
RUN mkdir /app/logs && chown ascii-art-web /app/logs

USER ascii-art-web

EXPOSE 8080

CMD [ "./ascii-art-web" ]