create table user
(
    id       bigint auto_increment
        primary key,
    uuid     varchar(36) not null,
    name     varchar(20) not null,
    password varchar(20) not null,
    question varchar(25) null,
    answer   varchar(25) null
)
    charset = utf8mb4;

