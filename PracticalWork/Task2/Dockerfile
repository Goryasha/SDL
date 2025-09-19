FROM postgres:16

RUN localedef -f UTF-8 -i ru_RU ru_RU.UTF-8

ARG POSTGRES_USER
ARG POSTGRES_PASSWORD

ADD DB_scripts/init.sql /docker-entrypoint-initdb.d/init.sql

RUN sed -i "s/POSTGRES_USER/$POSTGRES_USER/g" /docker-entrypoint-initdb.d/init.sql
RUN sed -i "s/POSTGRES_PASSWORD/\'$POSTGRES_PASSWORD\'/g" /docker-entrypoint-initdb.d/init.sql