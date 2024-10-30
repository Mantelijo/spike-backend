create table if not exists widgets(
    -- 10 million records will fit into serial, no need for bigserial
    id serial primary key,
    name text not null,
    serial_number text not null unique,
    -- null, q, r, p bits
    ports_bitmask bit(4) not null
);

-- serial_name might be used to retrieve the widgets, so we need an index for
-- serial_number strings.
create index widgets_serial_number on widgets using hash (serial_number);

-- Port type P connections
create table if not exists widget_connections(
    id serial primary key,
    widget_sn text not null unique,
    -- connection state can be not connected, therefore we allow nulls here.
    p_peer_sn text,
    r_peer_sn text,
    q_peer_sn text,
    
    -- make sure that the same connection is not added twice
    foreign key(widget_sn) references widgets(serial_number) on delete cascade,
    foreign key(p_peer_sn) references widgets(serial_number) on delete set null,
    foreign key(r_peer_sn) references widgets(serial_number) on delete set null,
    foreign key(q_peer_sn) references widgets(serial_number) on delete set null
);

create index widget_connections_widget_it on widget_connections using hash (widget_sn);
