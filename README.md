# htmlreport
Simple tool that generates HTML report for [hostchecker](https://github.com/h5vx/hostchecker), written in Go.
It uses `hosts.db` file and produce HTML file with hosts table. Table may be sorted by any column.

## Usage
Running without arguments, htmlreport will try to read `hosts.db` file in current working directory, then produce `hosts-report.html` at the same place. You may specify:
* **-dbpath** — path to DB file (including filename itself)
* **-o** — output file path
* **-server** (**NOT IMPLEMENTED YET**) — If this flag is set, start HTTP server and listen at specified interface:port, instead of generating file

## Requirements
* go-sqlite3 (install via `go get github.com/mattn/go-sqlite3`)
