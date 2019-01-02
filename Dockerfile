FROM ubuntu:18.10

MAINTAINER OlegSchwann

ENV PGVER 10
ENV GOVER 1.10

# Устанавливаем необходимые пакеты из репозиториев Ubuntu (пол года отставание от релиза).
# Обновляем список пакетов. Осторожно: Docker считает, что все функции
# детерминированны, и закеширует список пакетов намертво.
# Устанавливаем часовой пояс самостоятельно, избегаем попытки вывода
# красивого интерфейса выбора города во время сборки утилитой 'tzdata'.
ENV DEBIAN_FRONTEND 'noninteractive'
RUN echo 'Europe/Moscow' > '/etc/timezone' && \
    apt-get --yes update && \
    apt-get install --yes "postgresql-${PGVER}" "golang-${GOVER}" git;

# Выставляем переменные окружения сборки go проектов.
ENV GOROOT "/usr/lib/go-${GOVER}"
ENV GOPATH '/opt/go'
ENV PATH "${PATH}:${GOROOT}/bin:${GOPATH}/bin"

# Выкачиваем зависимости заранее, оптимизируя кеширование docker.
RUN go get \
    'github.com/buaazp/fasthttprouter' \
    'github.com/jackc/pgx' \
    'github.com/mailru/easyjson' \
    'github.com/pkg/errors' \
    'github.com/valyala/fasthttp' \
    'gopkg.in/inconshreveable/log15.v2';

# Копируем свой проект.
COPY . "${GOPATH}/src/github.com/OlegSchwann/2ch_api"

# Компилируем сервер.
RUN go build -o "${GOPATH}/bin/2ch_api" 'github.com/OlegSchwann/2ch_api'

# Кладём конфиг рядом с сервером. Пример конфига 'github.com/OlegSchwann/2ch_api/_build_configs/config.json'
RUN echo '{"DatabaseHost": "127.0.0.1", "DatabasePort": 5432, "DatabaseUser": "docker", "DatabasePassword": "docker", "DatabaseSpace": "docker", "ServerPort": 5000}' > "${GOPATH}/bin/config.json" && \
    echo "host all all 0.0.0.0/0 md5" >> "/etc/postgresql/${PGVER}/main/pg_hba.conf" && \
    printf "\n listen_addresses = '*' \n fsync = off \n synchronous_commit = off \n full_page_writes = off \n autovacuum = off \n wal_level = minimal \n max_wal_senders = 0 \n wal_writer_delay = 2000ms \n shared_buffers = 512MB \n effective_cache_size = 1024MB \n work_mem = 16MB \n" >> \
    "/etc/postgresql/${PGVER}/main/postgresql.conf"

# Публикуем порт сервера наружу.
EXPOSE 5000

USER postgres

RUN /etc/init.d/postgresql start &&\
    psql --echo-all --command "create user docker with superuser password 'docker';" &&\
    createdb -O docker docker &&\
    psql --dbname=docker --echo-all --command 'create extension if not exists "citext";' &&\
    /etc/init.d/postgresql stop;

# Запускаем PostgreSQL и сервер
CMD service postgresql start && 2ch_api "${GOPATH}/bin/config.json";
