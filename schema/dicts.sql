CREATE TABLE IF NOT EXISTS "dicts"
(
    "id"              integer      not null primary key autoincrement,
    "created_at"      datetime     not null,
    "updated_at"      datetime     not null,
    "word"            varchar(255) not null,
    "yomi"            varchar(255) not null,
    "changed_user_id" varchar(255) not null,
    "guild_id"        varchar(255) not null
);
