truncate quotes;

alter table quotes add "name" text not null;
alter table quotes add status text not null default 'new';

insert into quotes(symbol, name)
values ('AAPL', 'Apple Inc.'),
       ('F', 'Ford Motor Company'),
       ('T', 'AT&T Inc.'),
       ('NOK', 'Nokia Corporation'),
       ('XOM', 'Exxon Mobil Corporation'),
       ('MSFT', 'Microsoft Corporation'),
       ('TSLA', 'Tesla, Inc.'),
       ('UBER', 'Uber Technologies, Inc.'),
       ('INTC', 'Intel Corporation'),
       ('LYFT', 'Lyft, Inc.'),
       ('VZ', 'Verizon Communications Inc.'),
       ('SPCE', 'Virgin Galactic Holdings, Inc.'),
       ('GME', 'GameStop Corp.'),
       ('TWTR', 'Twitter, Inc.'),
       ('ZM', 'Zoom Video Communications, Inc.'),
       ('ORCL', 'Oracle Corporation'),
       ('FB', 'Facebook, Inc.'),
       ('QCOM', 'QUALCOMM Incorporated'),
       ('KO', 'The Coca-Cola Company'),
       ('BABA', 'Alibaba Group Holding Limited');
