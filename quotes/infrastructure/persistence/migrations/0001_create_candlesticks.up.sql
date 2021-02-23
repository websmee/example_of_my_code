create table candlesticks
(
    open      decimal not null,
    low       decimal not null,
    high      decimal not null,
    close     decimal not null,
    adj_close decimal not null,
    volume    bigint not null,
    timestamp timestamp not null,
    interval  text not null,
    quote_id  int not null,
    primary key (quote_id, interval, timestamp)
);

create index candlesticks_quote_id_interval_timestamp_idx on candlesticks(quote_id, interval, timestamp);