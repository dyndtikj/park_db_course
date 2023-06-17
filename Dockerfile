FROM golang:latest AS build

ADD . /app
WORKDIR /app
RUN CGO_ENABLED=0 go build ./cmd/main.go

FROM ubuntu:20.04
COPY . .

RUN apt-get -y update && apt-get install -y tzdata
RUN ln -snf /usr/share/zoneinfo/Russia/Moscow /etc/localtime && echo $TZ > /etc/timezone

ENV PGVER 12
RUN apt-get -y update && apt-get install -y postgresql-$PGVER
USER postgres

RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER dyndtikj WITH SUPERUSER PASSWORD 'postgres_pw';" &&\
    createdb -O dyndtikj forum &&\
    psql -f db/db.sql -d forum &&\
    /etc/init.d/postgresql stop

RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf
EXPOSE 5432
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

WORKDIR /usr/src/app

COPY . .
COPY --from=build /app/main .

USER root
EXPOSE 5000
CMD service postgresql start && ./main