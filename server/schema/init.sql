CREATE TABLE file
(
    id serial not null unique,
    uid integer not null unique ,
    path_to_file varchar(255) not null,
    chunk_size integer not null,
    meta text
);
