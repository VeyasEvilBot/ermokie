create table if not exists runs (
  id integer primary key autoincrement,
  name text not null,
  attempts integer default 0,
  active_split integer default 0,
  game text
    check (
      game is null or game in (
        "Demon's Souls",
        "Dark Souls 1",
        "Dark Souls 2",
        "Dark Souls 3",
        "Bloodborne",
        "Sekiro",
        "Elden Ring",
        "Marathon Slop",
        "Resident Evil 0",
        "Resident Evil",
        "Resident Evil 2",
        "Resident Evil 3",
        "Resident Evil 4",
        "Resident Evil 7",
        "Resident Evil 8",
        "Resident Evil Spinoffs",
        "Hades",
        "Cuphead",
        "Hollow Knight",
        "Hollow Knight: Silksong",
        "Celeste",
        "Crash Bandicoot",
        "Crash Bandicoot 2: Cortex Strikes Back",
        "Crash Bandicoot 3: Warped",
        "Crash Bandicoot 4: It’s About Time",
        "Ocarina of Time",
        "Majora's Mask",
        "Breath of the Wild",
        "Tears of the Kingdom",
        "Blasphemous",
        "Blasphemous 2",
        "Silent Hill",
        "Silent Hill 2 Remake",
        "Silent Hill 2 Classic",
        "Silent Hill 3",
        "Silent Hill 4",
        "Silent Hill Origins",
        "Dishonored",
        "Dishonored 2",
        "Dishonored: Death of the Outsider",
        "Tormented Souls",
        "Lies of P",
        "Oblivion Classic",
        "Oblivion Remastered",
        "Skyrim",
        "Minecraft",
        "Arkham Asylum",
        "Arkham Origins",
        "Arkham Knight",
        "Dead Cells",
        "Fallout 3",
        "Fallout: New Vegas",
        "Fallout 4",
        "The Binding of Isaac",
        "Ori and the Blind Forest",
        "Ori and the Will of the Wisps",
        "Clair Obscur: Expedition 33",
        "Thymesia",
        "Super Mario 64",
        "Super Mario Odyssey"
      )
    ),
  category text
    check (
      category is null or category in ("Any%", "All Bosses", "All Achievements", "All Great Runes")
    )
);

create table if not exists splits (
  id integer primary key autoincrement,
  run_id integer not null,
  name text not null,
  hit_count integer default 0 check (hit_count >= 0),
  pb_hit_count integer default 0 check (pb_hit_count >= 0),
  diff integer generated always as (
    case
      when coalesce(hit_count, 0) - coalesce(pb_hit_count, 0) < 0 then 0
      else coalesce(hit_count, 0) - coalesce(pb_hit_count, 0)
    end
  ) stored,
  idx integer not null,
  save_file text default null,
  foreign key (run_id) references runs(id) on delete cascade
);

create table if not exists misc_info (
  one integer not null default 1 check (one = 1) unique,
  active_run integer default null
);

insert or ignore into misc_info (one, active_run) values (1, null);

drop trigger if exists shift_splits_idx_before_insert;
create trigger shift_splits_idx_before_insert
before insert on splits
for each row
begin
  update splits
  set idx = idx + 1
  where run_id = NEW.run_id
    and idx >= NEW.idx;
end;

drop trigger if exists shift_splits_idx_after_delete;
create trigger shift_splits_idx_after_delete
after delete on splits
for each row
begin
  update splits
  set idx = idx - 1
  where run_id = OLD.run_id
    and idx > OLD.idx;
end;

drop trigger if exists reorder_splits_idx_after_update;
create trigger reorder_splits_idx_after_update
after update of idx on splits
for each row
when NEW.run_id = OLD.run_id and NEW.idx != OLD.idx
begin
  update splits
  set idx = idx - 1
  where run_id = NEW.run_id
    and idx > OLD.idx
    and idx <= NEW.idx;

  update splits
  set idx = idx + 1
  where run_id = NEW.run_id
    and idx >= NEW.idx
    and idx < OLD.idx;

end;
