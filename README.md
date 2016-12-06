# poke

poke is summoned for analyzing MySQL slow query logs, poke examines log
entries, converts text to JSON format and detects following features:

- time when query was started (using `Time:` and `Query_time` fields), 
- query type: `SELECT`, `INSERT`, `UPDATE`, `DELETE`, `DROP`
- query length

After analyzing slow logs poke will print JSON output to stdout, like as
following:

```
[
    {
        "bytes_sent": 0,
        "filesort": false,
        "filesort_on_disk": false,
        "full_join": false,
        "full_scan": false,
        "lock_time": 3e-05,
        "merge_passes": 0,
        "qc_hit": false,
        "query": "SET timestamp=1480443944;DELETE [...]",
        "query_length": 590,
        "query_time": 0.428312,
        "query_type": "DELETE",
        "rows_affected": 34,
        "rows_examined": 34,
        "rows_read": 0,
        "rows_sent": 0,
        "schema": "realty_multi",
        "time": "2016-11-30 01:25:44.00009175",
        "time_start": "2016-11-30 01:25:43.57177975",
        "tmp_disk_tables": 0,
        "tmp_table": false,
        "tmp_table_on_disk": false,
        "tmp_table_sizes": 0,
        "tmp_tables": 0
    },
    {
        "bytes_sent": 0,
        "filesort": false,
        "filesort_on_disk": false,
        "full_join": false,
        "full_scan": false,
        "lock_time": 3.6e-05,
        "merge_passes": 0,
        "qc_hit": false,
        "query": "SET timestamp=1480443944;INSERT [...]",
        "query_length": 560,
        "query_time": 0.374045,
        "query_type": "INSERT",
        "rows_affected": 31,
        "rows_examined": 31,
        "rows_read": 0,
        "rows_sent": 0,
        "schema": "realty_multi",
        "time": "2016-11-30 01:25:44.00072050",
        "time_start": "2016-11-30 01:25:43.62667550",
        "tmp_disk_tables": 0,
        "tmp_table": false,
        "tmp_table_on_disk": false,
        "tmp_table_sizes": 0,
        "tmp_tables": 0
    }
]
```

## What we can do with that JSON

We can use [github.com/kovetskiy/jsql](https://github.com/kovetskiy/jsql) and
query JSON dataset using SQL queries.


## Installation

Arch Linux User Repository:

[https://aur.archlinux.org/packages/jsql-git/](https://aur.archlinux.org/packages/jsql-git/)

or manually:

```
go get github.com/kovetskiy/jsql
```

## License
MIT.
