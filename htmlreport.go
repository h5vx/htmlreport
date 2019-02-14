package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// HostsEntry represents row from hosts database
type HostsEntry struct {
	Hostname string
	Type     string
	Reason   string
	Latency  float32
	Added    time.Time
	Updated  time.Time
}

// HostsTable contains fields that table template must know
type HostsTable struct {
	DateNow time.Time
	Hosts   []HostsEntry
}

// check that "-server" flag is specified (it may be empty)
func hasServerFlag() bool {
	for _, arg := range os.Args {
		if arg == "-server" || arg == "--server" ||
			strings.Contains(arg, "-server=") {
			return true
		}
	}
	return false
}

// FetchAllHosts fill hosts slice with data from hosts DB
func FetchAllHosts(hosts *[]HostsEntry, db *sql.DB) error {
	rows, err := db.Query(`
		SELECT hostname, type, reason, latency, added, updated 
		FROM hosts
	`)
	if err != nil {
		return fmt.Errorf("query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var host HostsEntry
		err := rows.Scan(&host.Hostname, &host.Type, &host.Reason,
			&host.Latency, &host.Added, &host.Updated)
		if err != nil {
			return fmt.Errorf("scan: %v", err)
		}

		*hosts = append(*hosts, host)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	return nil
}

// IsFileExists check that file is exists.
// May return true when some other error raised!
func IsFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}

func check(err error, errorFormatStr string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, errorFormatStr, err)
		os.Exit(1)
	}
}

func dbOpenOrDie(filename string) *sql.DB {
	if !IsFileExists(filename) {
		fmt.Fprintf(os.Stderr, "DB file \"%s\" is not exists\n", filename)
		os.Exit(-1)
	}
	db, err := sql.Open("sqlite3", filename)
	check(err, "while opening database: %v\n")
	return db
}

func main() {
	const templatePath = "template/table.html"
	var (
		dbPath = flag.String("dbpath", "hosts.db", "Path to hosts.db file")
		output = flag.String("o", "hosts-report.html", "Output file")
		listen = flag.String("server", ":8080", "If this flag is set, "+
			"start HTTP server and listen at\nspecified interface:port, "+
			"instead of generating file\nEvery HTTP request may "+
			"cause the database to reread")
	)
	flag.Parse()

	if hasServerFlag() {
		fmt.Fprintf(os.Stderr, "HTTP Server is not implemented yet!")
		fmt.Println("Must listen at", *listen)
	}

	db := dbOpenOrDie(*dbPath)
	defer db.Close()

	outputFile, err := os.OpenFile(*output,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0664) // rw-rw-r--
	check(err, "while opening output file: %v\n")
	defer outputFile.Close()

	tableTemplate, err := template.New("table.html").Funcs(template.FuncMap{
		"lower":   strings.ToLower,
		"timefmt": time.Time.Format,
		"isnulltime": func(t time.Time) bool {
			return t.Equal(time.Unix(0, 0))
		},
	}).ParseFiles(templatePath)
	check(err, "%v\n")

	hosts := make([]HostsEntry, 0, 1024)
	err = FetchAllHosts(&hosts, db)
	check(err, "while fetching hosts: %v\n")

	err = tableTemplate.Execute(outputFile, HostsTable{
		time.Now(),
		hosts,
	})
	check(err, "while execute template: %v\n")
}
