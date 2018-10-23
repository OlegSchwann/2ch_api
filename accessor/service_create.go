package accessor

import "github.com/jackc/pgx"

// Создаёт таблицы в базе данных. Нужно вызвать при подключении.
func init() {
	Prep.addFirst(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
begin transaction;
-- подключаем расширение с регистронезависимым текстом
create extension if not exists "citext";

-- Пользователь.
create table if not exists "user" (
  "nickname" citext primary key,
    -- Имя пользователя (уникальное поле).
    -- Данное поле допускает только латиницу, цифры и знак подчеркивания.
    -- Сравнение имени регистронезависимо (индекс "user_nickname_lower_unique").
    -- Не должно изменяться при обновлении.
  "about"    text null,
    -- Описание пользователя.
  "email"    citext not null unique,
    -- Почтовый адрес пользователя.
  "fullname" text not null
    -- Полное имя пользователя.
);

-- Форум
create table if not exists "forum" (
  "slug"  citext primary key,
    -- https://ru.wikipedia.org/wiki/Семантический_URL
  "title" text not null,
    -- Название форума.
  "user"  citext not null references "user" ("nickname")
);

-- Ветка обсуждения на форуме.
create table if not exists "thread" (
  "id"      serial4 primary key,
    -- Идентификатор ветки обсуждения.
  "author"  citext not null,
    -- Пользователь, создавший данную тему.
  "created" timestamp with time zone,
    -- Дата создания ветки на форуме.
  "forum"   citext references "forum" ("slug"),
    -- Форум, в котором расположена данная ветка обсуждения.
  "message" text not null,
    -- Описание ветки обсуждения.
  "slug"    citext null,
    -- https://ru.wikipedia.org/wiki/Семантический_URL, опциональная строка.
  "title"   text not null
    -- Заголовок ветки обсуждения.
);

-- Информация о голосовании пользователя.
create table if not exists "vote" (
  "nickname" citext     not null references "user" ("nickname"),
    -- Идентификатор пользователя.
  "voice"    smallint not null,
    -- Отданный голос ∈ [1, -1].
  "id"       integer  not null references "thread" ("id")
    -- Идентификатор ветки обсуждения на форуме.
);

-- Посты в ветке обсуждения на форуме.
create table if not exists "post" (
  "thread"             integer                  not null references "thread" ("id"),
    -- Идентификатор ветки (id) обсуждения.
  "author"             citext                     not null references "user" ("nickname"),
    -- Автор, написавший данное сообщение.
  "created"            timestamp with time zone not null,
    -- Дата создания сообщения на форуме.
  "id"                 serial8 primary key,
    -- Идентификатор данного сообщения.
  "is_edited"          boolean                  not null,
    -- Истина, если данное сообщение было изменено.
  "message"            text                     not null,
    -- Собственно сообщение форума.
  "parent"             bigint                   not null,
    -- Идентификатор родительского сообщения (0 - корневое сообщение обсуждения).

-- Нижнее не нормализовано:
  "forum"              citext references "forum" ("slug"),
    -- Идентификатор форума (slug) данного сообещния.
  "materialized_path"  text                     not null,
    -- Материализованный путь в дереве сообщений.
  "number_of_children" integer                  not null
    -- Количество детей для быстрого вычисления материализованного пути потомка.
);

commit;

/* Структура для быстрого нахождения дерева коммкетариев:
┌─────────┬──────────────────────┬────────┐
│структура│Материализованный путь│потомков│
├─────────┼──────────────────────┼────────┤
│ a       │ 01                   │ 2      │
│ -b      │ 01.01                │ 1      │
│ --c     │ 01.01.01             │ 0      │
│ -d      │ 01.02                │ 3      │
│ --e     │ 01.02.01             │ 1      │
│ ---f    │ 01.02.01.01          │ 0      │
│ --g     │ 01.02.02             │ 1      │
│ ---h    │ 01.02.02.01          │ 0      │
│ --$     │ 01.02.03             │ 0      │
│ i       │ 02                   │ 1      │
│ -j      │ 02.01                │ 0      │
│ k       │ 03                   │ 1      │
│ -l      │ 03.01                │ 0      │
│ m       │ 04                   │ 1      │
│ -n      │ 04.01                │ 3      │
│ --o     │ 04.01.01             │ 1      │
│ ---p    │ 04.01.01.01          │ 1      │
│ ----q   │ 04.01.01.01.01       │ 1      │
│ -----r  │ 04.01.01.01.01.01    │ 1      │
│ ------s │ 04.01.01.01.01.01.01 │ 1      │
│ -t      │ 04.02                │ 2      │
│ --u     │ 04.02.01             │ 1      │
│ ---v    │ 04.02.01.01          │ 0      │
│ --w     │ 04.02.02             │ 0      │
│ -x      │ 04.03                │ 0      │
│ y       │ 05                   │ 0      │
│ z       │ 06                   │ 0      │
└─────────┴──────────────────────┴────────┘
*/`
		_, err = conn.Exec(sql)
		return
	})
}
