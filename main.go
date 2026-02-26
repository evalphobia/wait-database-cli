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
	_ "github.com/lib/pq"
	"golang.org/x/net/proxy"
)

const (
	timeout       = 600 // ≒ 10 minutes
	logInterval   = 5   // ≒ 5 seconds
	graceInterval = 30  // grace period after startup (retry on all errors)
)

func main() {
	var dbType string
	var host string
	var port int
	var user string
	var password string
	var database string
	var query string
	flag.StringVar(&dbType, "type", os.Getenv("WAIT_DATABASE_TYPE"), "Database type (mysql or postgresql)")
	flag.StringVar(&host, "host", os.Getenv("WAIT_DATABASE_HOST"), "Database host")
	flag.IntVar(&port, "port", 0, "Database port")
	flag.StringVar(&user, "user", os.Getenv("WAIT_DATABASE_USER"), "Database user")
	flag.StringVar(&password, "password", os.Getenv("WAIT_DATABASE_PASSWORD"), "Database password")
	flag.StringVar(&database, "database", os.Getenv("WAIT_DATABASE_DATABASE"), "Database name")
	flag.StringVar(&query, "query", os.Getenv("WAIT_DATABASE_QUERY"), "SQL to execute")
	flag.Parse()

	// Set default database type
	if dbType == "" {
		dbType = "mysql"
	}

	// Validate database type
	if dbType != "mysql" && dbType != "postgresql" {
		log.Fatalf("Invalid database type: %s. Must be either 'mysql' or 'postgresql'", dbType)
	}

	// Set default port based on database type
	if port == 0 {
		portEnv := os.Getenv("WAIT_DATABASE_PORT")
		if portEnv != "" {
			p, err := strconv.Atoi(portEnv)
			if err != nil {
				log.Fatalf("Invalid port number: %s", portEnv)
			}
			port = p
		} else {
			if dbType == "mysql" {
				port = 3306
			} else {
				port = 5432
			}
		}
	}

	// Set default user
	if user == "" {
		if dbType == "mysql" {
			user = "root"
		} else {
			user = "postgres"
		}
	}

	if database == "" && query != "" {
		log.Fatal("Please specify the database name with the -database option when using the -query option.")
	}

	var db *sql.DB
	var err error

	if dbType == "mysql" {
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

		db, err = sql.Open("mysql", dns)
	} else {
		// PostgreSQL connection
		dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
			host, port, user, database)
		if password != "" {
			dsn += fmt.Sprintf(" password=%s", password)
		}

		db, err = sql.Open("postgres", dsn)
	}

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	i := 0
	for {
		i++
		if i > timeout {
			log.Fatalf("[%ds] %s connection timeout error\n", i, dbType)
		}

		err = db.Ping()
		if err != nil {
			// after grace period, only retry on known waiting errors
			if i > graceInterval && !isWaitingError(err, dbType) {
				panic(err)
			}
			if i%logInterval == 0 {
				log.Printf("[%ds] Waiting for %s to start...(status: %+v)\n ", i, dbType, err)
			}
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	if query != "" {
		log.Printf("Query: %q\n", query)
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
				row = append(row, fmt.Sprintf("%s:%v", colName, raw_value))
			}
			log.Printf("row[%d]: %s\n", num, strings.Join(row, ", "))
		}
	}
}

func isWaitingError(err error, dbType string) bool {
	e := err.Error()
	switch {
	case strings.HasSuffix(e, "connect: connection refused"):
		return true
	case dbType == "postgresql" && strings.Contains(e, "connection refused"):
		return true
	case dbType == "postgresql" && strings.Contains(e, "no such host"):
		return true
	case dbType == "postgresql" && strings.Contains(e, "connection reset by peer"):
		return true
	}
	return false
}
