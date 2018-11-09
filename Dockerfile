FROM ubuntu:18.10

MAINTAINER OlegSchwann

ENV PGVER 10
ENV GOVER 1.10

# Устанавливаем необходимые пакеты из репозиториев Ubuntu (пол года отставание от релиза).
# Обновляем список пакетов. Осторожно: Docker считает, что все функции
# детерминированны, и закеширует список пакетов намертво.
RUN apt-get -y update

# устанавливаем часовой пояс самостоятельно, избегаем попытки вывода
# красивого интерфейса выбора города во время сборки утилитой 'tzdata'.
ENV DEBIAN_FRONTEND 'noninteractive'
RUN echo 'Europe/Moscow' > '/etc/timezone'

RUN apt-get install --yes "postgresql-${PGVER}"
RUN apt install -y "golang-${GOVER}" git

# Выставляем переменные окружения сборки go проектов.
ENV GOROOT "/usr/lib/go-${GOVER}"
ENV GOPATH '/opt/go'
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    ENV PATH "${GOROOT}/bin:${GOPATH}/bin:/usr/local/go/bin:${PATH}"
# Выкачиваем зависимости и компилируем сервер.
RUN go get 'github.com/OlegSchwann/2ch_api';
RUN go build -o "${GOPATH}/bin/2ch_api" 'github.com/OlegSchwann/2ch_api'

# Кладём конфиг рядом с сервером. Пример конфига 'github.com/OlegSchwann/2ch_api/_build_configs/config.json'
RUN echo '{"DatabaseHost": "127.0.0.1", "DatabasePort": 5432, "DatabaseUser": "docker", "DatabasePassword": "docker", "DatabaseSpace": "docker", "ServerPort": 5000}' > "${GOPATH}/bin/config.json";
RUN echo "host all all 0.0.0.0/0 md5" >> "/etc/postgresql/${PGVER}/main/pg_hba.conf"

RUN grep --perl-regexp '^\s*#|^\s*$' --invert-match "/etc/postgresql/${PGVER}/main/postgresql.conf"

# Перезатираем конфиг базы: там отключение синхронного комита и другие интересные вещи.
RUN mv --force "${GOPATH}/src/github.com/OlegSchwann/2ch_api/_build_configs/postgresql.conf" "/etc/postgresql/${PGVER}/main/postgresql.conf"

# Публикуем порт сервера наружу.
EXPOSE 5000

USER postgres

RUN /etc/init.d/postgresql start &&\
    psql --echo-all --command "create user docker with superuser password 'docker';" &&\
    createdb -O docker docker &&\
    psql --dbname=docker --echo-all --command 'create extension if not exists "citext";' &&\
    /etc/init.d/postgresql stop;

# -- EXPOSE 5432

# Запускаем PostgreSQL и сервер
CMD service postgresql start && 2ch_api "${GOPATH}/bin/config.json";
