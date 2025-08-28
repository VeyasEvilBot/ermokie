create table if not exists runs(
	id integer primary key autoincrement,
	name text not null,
	attempts integer default 0,
	active_split ingeger default 0
);

create table if not exists splits(
	id integer primary key autoincrement,
	run_id integer not null,
	name text not null,
	hit_count integer default 0,
	pb_hit_count integer default 0,
	idx integer not null,
	foreign key (run_id) references runs(id) on delete cascade
);
