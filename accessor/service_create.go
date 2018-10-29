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
  "number_of_children" integer                  not null default 0
    -- количество постов первого уровня в цепочке обсуждений. 
);

create unique index if not exists "thread_slug_unique" on "thread"("slug") where "slug" <> '';

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

-- new (вставляемая строка) гарантированно содержит:
--   "thread_id" или "thread_slug"
--   "author"
--   "created" -- вставляется на стороне application server
--   "message"  
--   "parent" -- может быть равно 0, если сообщение корневое

-- не нуждаются в триггере
--   "id" -- добавляется самостоятельно из serial
--   "is_edited" -- по умолчанию false
	
-- поддерживаются триггером:
--   "forum" -- берётся из "thread"."forum" 
--   "materialized_path"
--   "number_of_children" -- важно обновить для родительского сообщения.

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
*/

create or replace function "post_consistency_support"() returns trigger as $$
declare
  "parent_number_of_children" integer;
  "parent_materialized_path" text;
begin
  -- Восстанавливаем "thread_id" по "thread_slug" или "thread_slug" по "thread_id"
  if (new."thread_id" <> 0) then
    select "forum", "slug"
    into new."forum", new."thread_slug"
    from "thread"
    where "id" = new."thread_id";

    if (new."forum" is null) then
      raise exception sqlstate '23503' using message = 'can not find "thread" where "id" = ' || new."thread_id";
    end if;
  else -- if (new."thread_slug" <> '')
    select "forum", "id"
    into new."forum", new."thread_id"
    from "thread"
    where "slug" = new."thread_slug";

    if (new."forum" is null) then
      raise exception sqlstate '23503' using message = 'can not find "thread" where "slug" = ' || new."thread_slug";
    end if;
  end if;

  -- Находим и инкрементируем количество детей у родительских элементов.
  -- Осуществляем сборку материализованного пути.
  if (new."parent" = 0) then
    update "thread"
    set "number_of_children" = "number_of_children" + 1
    where "id" = new."thread_id"
    returning "number_of_children"
    into "parent_number_of_children";

    if ("parent_number_of_children" is null) then
      -- поддерживаем целостность внешнего ключа - падаем, если не нашли родителя.
      raise exception sqlstate '23503' using message = 'can not find "post" where "parent" = ' || new."parent";
    end if;

    new."materialized_path" = to_char("parent_number_of_children", 'FM0000');
  else
    update "post"
    set "number_of_children" = "number_of_children" + 1
    where "id" = new."parent"
    returning "number_of_children", "materialized_path"
    into "parent_number_of_children", "parent_materialized_path";

    if ("parent_number_of_children" is null) then
      -- поддерживаем целостность внешнего ключа - падаем, если не нашли родителя.
      raise exception sqlstate '23503' using message = 'can not find "post" where "parent" = ' || new."parent";
    end if;

    new."materialized_path" = concat("parent_materialized_path", '.', to_char("parent_number_of_children", 'FM0000'));
  end if;
  return new;
end;
$$ language plpgsql;

drop trigger if exists "post_consistency_support_trigger" on "post";

create trigger "post_consistency_support_trigger"
  before insert on "post"
  for each row
execute procedure "post_consistency_support"();

commit;`
		_, err = conn.Exec(sql)
		return
	})
}
