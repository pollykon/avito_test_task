create table "user"(
    id bigint primary key
);

create table segment(
    id text primary key,
    deleted bool not null default false
);

create table user_segment(
    user_id bigint references "user"(id),
    segment_id text references segment(id),
    primary key (user_id, segment_id),
    insert_time timestamp with time zone default now() not null,
    ttl interval
);

create table log(
    log_id bigserial primary key,
    user_id bigint,
    segment_id text,
    operation text,
    insert_time timestamp with time zone default now() not null
);

create index log_user_id_insert_time_ix on log(user_id, insert_time desc);
