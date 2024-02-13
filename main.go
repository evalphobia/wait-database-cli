package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/net/proxy"
)

const (
	timeout     = 600 // ≒ 10 minutes
	logInterval = 5   // ≒ 5 seconds
)

func main() {
	var host string
	var port int
	var user string
	var password string
	var database string
	var query string
	flag.StringVar(&host, "host", os.Getenv("WAIT_MYSQL_HOST"), "MySQL host")
	flag.IntVar(&port, "port", 3306, "MySQL port")
	flag.StringVar(&user, "user", os.Getenv("WAIT_MYSQL_USER"), "MySQL user")
	flag.StringVar(&password, "password", os.Getenv("WAIT_MYSQL_PASSWORD"), "MySQL password")
	flag.StringVar(&database, "database", os.Getenv("WAIT_MYSQL_DATABASE"), "MySQL database")
	flag.StringVar(&query, "query", os.Getenv("WAIT_MYSQL_QUERY"), "SQL to execute")
	flag.Parse()

	if user == "" {
		user = "root"
	}
	portEnv := os.Getenv("WAIT_MYSQL_PORT")
	if port != 3306 && portEnv != "" {
		p, err := strconv.Atoi(portEnv)
		if err != nil {
			log.Fatalf("Invalid port number: %s", portEnv)
		}
		port = p
	}

	if database == "" && query != "" {
		log.Fatal("Please specify the database name with the -database option when using the -query option.")
	}

	// for socks proxy (e.g. ALL_PROXY=socks5://localhost:9999/)
	dialer := proxy.FromEnvironment()
	mysql.RegisterDialContext("tcp", func(ctx context.Context, network string) (net.Conn, error) {
		return dialer.Dial("tcp", network)
	})

	dns := ""
	switch {
	case password == "":
		dns = fmt.Sprintf("%s@tcp(%s:%d)/%s", user, host, port, database)
	default:
		dns = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, database)
	}

	db, err := sql.Open("mysql", dns)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	i := 0
	for {
		i++
		if i > timeout {
			log.Fatalf("[%ds] MySQL connection timeout error\n", i)
		}

		err = db.Ping()
		if err != nil {
			if !isWaitingError(err) {
				panic(err)
			}
			if i%logInterval == 0 {
				log.Printf("[%ds] Waiting for MySQL to start...(status: %+v)\n ", i, err)
			}
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	if query != "" {
		log.Printf("Query: '%s'\n", query)
		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			log.Fatal(err)
		}

		num := 0
		for rows.Next() {
			num++
			values := make([]interface{}, len(columns))
			for i := range values {
				var v interface{}
				values[i] = &v
			}
			if err := rows.Scan(values...); err != nil {
				log.Fatal(err)
			}
			row := make([]string, 0, len(columns))
			for i, colName := range columns {
				raw_value := *(values[i].(*interface{}))
				row = append(row, fmt.Sprintf("%s:%s", colName, raw_value))
			}
			log.Printf("row[%d]: %s\n", num, strings.Join(row, ", "))
		}
	}
}

func isWaitingError(err error) bool {
	e := err.Error()
	switch {
	case strings.HasSuffix(e, "connect: connection refused"):
		return true
	}
	return false
}
