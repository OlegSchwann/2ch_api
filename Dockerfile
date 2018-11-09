FROM ubuntu:18.10

MAINTAINER OlegSchwann

# Устанавливаем необходимые пакеты из репозиториев Ubuntu (пол года отставание от релиза).
ENV PGSQL_VERSION '10'
ENV GO_VERSION '1.10'

# Обновляем список пакетов. Осторожно: Docker считает, что все функции
# детерминированны, и закеширует список пакетов намертво.
RUN apt-get --yes update;

# устанавливаем часовой пояс самостоятельно, избегаем попытки вывода
# красивого интерфейса выбора города во время сборки утилитой 'tzdata'.
ENV DEBIAN_FRONTEND 'noninteractive'
RUN echo 'Europe/Moscow' > '/etc/timezone' && \
    apt-get install --yes "postgresql-${PGSQL_VERSION}" "golang-${GO_VERSION}" git ;

# Выставляем переменные окружения сборки go проектов.
ENV GOROOT "/usr/lib/go-${GO_VERSION}"
ENV GOPATH '/opt/go'
ENV PATH "${PATH}:${GOROOT}/bin:${GOPATH}/bin:/usr/local/go/bin"

# Копируем код и докачиваем зависимости.
RUN go get 'github.com/OlegSchwann/2ch_api'

# Компилируем сервер
RUN go build -o "${GOPATH}/bin/2ch_api" 'github.com/OlegSchwann/2ch_api'

# Подкладываем конфиг в папку с сервером.
RUN echo '{"DatabaseHost": "127.0.0.1", "DatabasePort": 5432, "DatabaseUser": "server", "DatabasePassword": "", "DatabaseSpace": "server", "ServerPort": 5000}' > "${GOPATH}/bin/config.json";

# Заставляем PostgreSQL принимать соединения отовсюду.
RUN echo "host all all 0.0.0.0/0 trust" >> "/etc/postgresql/${PGSQL_VERSION}/main/pg_hba.conf"

# Перезатираем конфиг базы: там отключение синхронного комита и другие интересные вещи.
RUN mv --force "${GOPATH}/src/github.com/OlegSchwann/2ch_api/_build_configs/postgresql.conf" "/etc/postgresql/${PGSQL_VERSION}/main/postgresql.conf"

# Run the rest of the commands as the ``postgres`` user created by the ``postgres-$PGVER`` package when it was ``apt-get installed``
USER postgres

# Create a PostgreSQL role named 'server' with password and
# then create a database `docker` owned by the ``docker`` role.
RUN /etc/init.d/postgresql start && \
    psql --dbname='server' --no-password --host=127.0.0.1 --port=5432 --command='create user "server" with superuser password "server";' && \
    /etc/init.d/postgresql stop ;
#    createdb -O docker docker && \

USER root

# Публикуем порт сервера наружу.
EXPOSE 5000

# Запускаем PostgreSQL и сервер
CMD service postgresql start && 2ch_api ;
