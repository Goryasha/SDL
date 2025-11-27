FROM postgres:16

RUN localedef -f UTF-8 -i ru_RU ru_RU.UTF-8

ARG POSTGRES_DB
ARG POSTGRES_SCHEMA

ARG CUSTOM_USER
ARG CUSTOM_PASSWORD

ADD DB_scripts/init.sql /docker-entrypoint-initdb.d/init.sql

RUN sed -i "s/CUSTOM_USER/$CUSTOM_USER/g" /docker-entrypoint-initdb.d/init.sql
RUN sed -i "s/CUSTOM_PASSWORD/\'$CUSTOM_PASSWORD\'/g" /docker-entrypoint-initdb.d/init.sql

RUN sed -i "s/POSTGRES_DB/$POSTGRES_DB/g" /docker-entrypoint-initdb.d/init.sql
RUN sed -i "s/POSTGRES_SCHEMA/$POSTGRES_SCHEMA/g" /docker-entrypoint-initdb.d/init.sql

ADD DB_scripts/setup_data.sql /docker-entrypoint-initdb.d/setup_data.sql

RUN sed -i "s/POSTGRES_DB/$POSTGRES_DB/g" /docker-entrypoint-initdb.d/setup_data.sql
RUN sed -i "s/POSTGRES_SCHEMA/$POSTGRES_SCHEMA/g" /docker-entrypoint-initdb.d/setup_data.sql