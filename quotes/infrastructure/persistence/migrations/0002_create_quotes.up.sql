create table quotes
(
    id serial not null primary key,
    symbol text not null
);

insert into quotes(id, symbol) values (1, 'GC=F')