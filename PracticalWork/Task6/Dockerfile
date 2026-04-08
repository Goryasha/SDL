FROM postgres:16

RUN localedef -f UTF-8 -i ru_RU ru_RU.UTF-8

ADD DB_scripts/init_filled.sql /docker-entrypoint-initdb.d/init.sql
