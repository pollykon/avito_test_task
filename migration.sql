create table "user"(
    id bigint primary key
);

create table segment(
    id text primary key,
    deleted bool not null default false,
    percent bigint check ( 0 < percent and percent <= 100)
);

create table log(
    id bigserial primary key,
    user_id bigint,
    segment_id text,
    operation text,
    insert_time timestamp with time zone default now() not null
);

create table user_segment(
     id bigserial primary key,
     user_id bigint references "user"(id),
     segment_id text references segment(id),
     insert_time timestamp with time zone default now() not null,
     ttl interval
);

create index log_user_id_insert_time_ix on log(user_id, insert_time desc);
