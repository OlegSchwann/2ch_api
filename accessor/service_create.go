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
  "about"    text   not null,
    -- Описание пользователя. Необязательное, может быть равно ''.
  "email"    citext not null unique,
    -- Почтовый адрес пользователя.
  "fullname" text   not null
    -- Полное имя пользователя.
);

-- Форум
create table if not exists "forum" (
  "slug"  citext primary key,
    -- https://ru.wikipedia.org/wiki/Семантический_URL
  "title" text   not null,
    -- Название форума.
  "user"  citext not null references "user" ("nickname")
);

-- Ветка обсуждения на форуме.
create table if not exists "thread" (
  "author"             citext                   not null references "user"("nickname"),
    -- Пользователь, создавший данную тему.
  "created"            timestamp with time zone not null,
    -- Дата создания ветки на форуме.
  "forum"              citext                   not null references "forum" ("slug"),
    -- Форум, в котором расположена данная ветка обсуждения.
  "id"                 serial4                  primary key,
    -- Идентификатор ветки обсуждения.
  "message"            text                     not null,
    -- Описание ветки обсуждения.
  "slug"               citext                   not null,
    -- https://ru.wikipedia.org/wiki/Семантический_URL, опциональная строка, может быть равна ''.
  "title"              text                     not null,
    -- Заголовок ветки обсуждения.

-- Не нормализовано.
  "votes"              integer                  not null default 0,
    -- суммарное количество голосов за и против треда. Подерживается триггером при изменении в "vote".
  "number_of_children" integer                  not null default 0
    -- количество постов первого уровня в цепочке обсуждений. 
);

create unique index if not exists "thread_slug_unique" on "thread"("slug") where "slug" <> '';

-- Информация о голосовании пользователя.
create table if not exists "vote" (
  "nickname"  citext   not null references "user" ("nickname"),
    -- Идентификатор пользователя.
  "voice"     smallint not null,
    -- Отданный голос ∈ [1, -1].
  "thread_id" integer  not null references "thread" ("id")
    -- Идентификатор ветки обсуждения на форуме.
);

-- Каждый пользователь за один тред может проголосовать не более одного раза, но может изменить своё мнение.
create unique index if not exists "vote_nickname_id_unique" on "vote" ("nickname", "thread_id"); 

-- Посты в ветке обсуждения на форуме.
create table if not exists "post" (
  "thread_id"          integer                  not null references "thread" ("id"),
    -- Идентификатор ветки (id) обсуждения.
  "author"             citext                   not null references "user" ("nickname"),
    -- Автор, написавший данное сообщение.
  "created"            timestamp with time zone not null,
    -- Дата создания сообщения на форуме.
  "id"                 serial8 primary key,
    -- Идентификатор данного сообщения.
  "is_edited"          boolean                  not null default false,
    -- Истина, если данное сообщение было изменено.
  "message"            text                     not null,
    -- Собственно сообщение форума.
  "parent"             bigint                   not null,
    -- Идентификатор родительского сообщения (0 - корневое сообщение обсуждения).
    -- references "post" ("id") сделать нельзя из-за 0.

-- Не нормализованые поля, поддерживаются триггером:
  "forum"              citext                   not null references "forum" ("slug"),
    -- Идентификатор форума (slug) данного сообещния.
  "thread_slug"        citext                   not null,  -- references "thread" (slug) нельзя сделать через sql - slug не уникаен, есть множество slug = ''
    -- Идентификатор ветки (slug) обсуждения, опционально, может быть равна ''.
    -- Однако запрос может содержать "thread_slug", а не "thread_id", триггером надо восстановить последнее.
  "materialized_path"  text                     not null,
    -- Материализованный путь в дереве сообщений, структура ниже.
  "number_of_children" integer                  not null default 0
    -- Количество детей для быстрого вычисления материализованного пути потомка.
);

-- оптимизация префиксного сравнения select * from "post" where "thread_id" = 1 && "materialized_path" => '0001' && "materialized_path" < '0003';
-- отлавливает неверные материализованные пути, появляющиеся при гонке данных.
create unique index if not exists "post_materialized_path" on "post" ("thread_id",   "materialized_path");

-- для сортировки комментариев по дате (flat)
create unique index if not exists "post_materialized_path" on "post" ("thread_id", "created");


/* Структура для быстрого нахождения дерева коммкетариев:
количество потомков - количество реальных потомков, первоначально равно 0.
отсчёт материализованного пути начинается с 1, количество потомков всегда рано номеру последнего потомка
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
*/

commit;`
		_, err = conn.Exec(sql)
		return
	})
}
