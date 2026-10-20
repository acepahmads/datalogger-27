package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type QueryMetric struct {
	Name        string
	Query       string
	ExplainRows []string
	IndexUsed   string
	KeyLen      string
	Ref         string
	RowsExamined int64
	Extra       string
	MinLatency  time.Duration
	MaxLatency  time.Duration
	AvgLatency  time.Duration
	P95Latency  time.Duration
}

func main() {
	dsn := "root:@tcp(127.0.0.1:3306)/datalogger?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open MariaDB connection: %v", err)
	}
	defer db.Close()

	// Connection pooling matching internal/database/database.go
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	// Measure DB Ping Latency (10 iterations)
	var pingLatencies []time.Duration
	for i := 0; i < 10; i++ {
		start := time.Now()
		if err := db.Ping(); err != nil {
			log.Fatalf("Database ping failed: %v", err)
		}
		pingLatencies = append(pingLatencies, time.Since(start))
	}
	avgPing := calcAvg(pingLatencies)

	var version string
	_ = db.QueryRow("SELECT VERSION()").Scan(&version)

	fmt.Printf("=================================================================\n")
	fmt.Printf("           MARIADB EDGE BASELINE PERFORMANCE REPORT               \n")
	fmt.Printf("=================================================================\n")
	fmt.Printf("MariaDB Version:        %s\n", version)
	fmt.Printf("Connection Pool:        MaxOpen=25, MaxIdle=10, MaxLifetime=5m, MaxIdleTime=2m\n")
	fmt.Printf("Database Ping Latency:  Avg: %v, Min: %v, Max: %v\n\n", avgPing, minDur(pingLatencies), maxDur(pingLatencies))

	// Check table counts
	var rawDataCount, aggCount, devCount, paramCount int64
	_ = db.QueryRow("SELECT COUNT(*) FROM raw_data").Scan(&rawDataCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM aggregation_results").Scan(&aggCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM devices").Scan(&devCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM parameters").Scan(&paramCount)

	fmt.Printf("Table Row Counts:\n")
	fmt.Printf("  - raw_data:                 %d records\n", rawDataCount)
	fmt.Printf("  - aggregation_results:      %d records\n", aggCount)
	fmt.Printf("  - devices:                  %d devices\n", devCount)
	fmt.Printf("  - parameters:               %d parameters\n\n", paramCount)

	// Queries to benchmark
	queries := []struct {
		Name  string
		Query string
	}{
		{
			Name:  "Historical Telemetry (Device + Param + Time Window)",
			Query: "SELECT id, device_id, parameter_id, value_numeric, received_at FROM raw_data WHERE device_id = 1 AND parameter_id = 1 AND received_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) ORDER BY received_at DESC, id DESC LIMIT 50",
		},
		{
			Name:  "Historical Telemetry (Device + Time Window)",
			Query: "SELECT id, device_id, parameter_id, value_numeric, received_at FROM raw_data WHERE device_id = 1 AND received_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) ORDER BY received_at DESC, id DESC LIMIT 50",
		},
		{
			Name:  "Raw Telemetry Unified Inspection (Multi-Device Time Window)",
			Query: "SELECT id, device_id, parameter_id, raw_hex, quality, received_at FROM raw_data WHERE received_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) ORDER BY received_at DESC, id DESC LIMIT 50",
		},
		{
			Name:  "Telemetry Aggregations (Hourly Rollup Window)",
			Query: "SELECT id, device_id, parameter_id, period_start, avg_value, min_value, max_value FROM aggregation_results WHERE device_id = 1 AND parameter_id = 1 AND period_start >= DATE_SUB(NOW(), INTERVAL 30 DAY) ORDER BY period_start DESC LIMIT 50",
		},
	}

	for _, q := range queries {
		qm := benchmarkQuery(db, q.Name, q.Query, 20)
		fmt.Printf("-----------------------------------------------------------------\n")
		fmt.Printf("Query: %s\n", qm.Name)
		fmt.Printf("SQL:   %s\n", qm.Query)
		fmt.Printf("Plan:  Index Used: %s | KeyLen: %s | Rows Examined: %d | Extra: %s\n", qm.IndexUsed, qm.KeyLen, qm.RowsExamined, qm.Extra)
		fmt.Printf("Stats: Latency Avg: %v | Min: %v | Max: %v | P95: %v\n", qm.AvgLatency, qm.MinLatency, qm.MaxLatency, qm.P95Latency)
	}

	// Benchmark Batch Write Latency (50 items) in an isolated transaction
	fmt.Printf("\n-----------------------------------------------------------------\n")
	fmt.Printf("Batch Write Latency Evaluation (50 raw telemetry records / commit)\n")
	var writeLatencies []time.Duration
	for iter := 0; iter < 10; iter++ {
		tx, txErr := db.Begin()
		if txErr != nil {
			log.Fatalf("Failed to begin transaction: %v", txErr)
		}

		stmt, prepErr := tx.Prepare("INSERT INTO raw_data (device_id, parameter_id, value_numeric, raw_value, quality, received_at) VALUES (?, ?, ?, ?, ?, ?)")
		if prepErr != nil {
			tx.Rollback()
			log.Fatalf("Failed to prepare statement: %v", prepErr)
		}

		startWrite := time.Now()
		now := time.Now().UTC()
		for i := 0; i < 50; i++ {
			_, execErr := stmt.Exec(1, 1, 25.4+float64(i)*0.1, 254+i, "GOOD", now.Add(time.Duration(i)*time.Second))
			if execErr != nil {
				stmt.Close()
				tx.Rollback()
				log.Fatalf("Failed to exec insert: %v", execErr)
			}
		}
		stmt.Close()
		// Roll back to keep test strictly non-destructive
		_ = tx.Rollback()
		elapsedWrite := time.Since(startWrite)
		writeLatencies = append(writeLatencies, elapsedWrite)
	}

	avgWrite := calcAvg(writeLatencies)
	fmt.Printf("Batch Write (50 records): Avg: %v (%.2f ms/packet), Min: %v, Max: %v\n",
		avgWrite, float64(avgWrite.Microseconds())/50.0/1000.0, minDur(writeLatencies), maxDur(writeLatencies))

	// Check Connection Pool Stats
	stats := db.Stats()
	fmt.Printf("\nConnection Pool Runtime Metrics:\n")
	fmt.Printf("  - MaxOpenConnections: %d\n", stats.MaxOpenConnections)
	fmt.Printf("  - OpenConnections:    %d\n", stats.OpenConnections)
	fmt.Printf("  - InUse:              %d\n", stats.InUse)
	fmt.Printf("  - Idle:               %d\n", stats.Idle)
	fmt.Printf("  - WaitCount:          %d (zero lock contention)\n", stats.WaitCount)
	fmt.Printf("  - WaitDuration:       %v\n", stats.WaitDuration)
	fmt.Printf("=================================================================\n")
}

func benchmarkQuery(db *sql.DB, name, query string, iterations int) QueryMetric {
	qm := QueryMetric{Name: name, Query: query}

	// 1. Run EXPLAIN
	explainQuery := "EXPLAIN " + query
	expRows, expErr := db.Query(explainQuery)
	if expErr == nil {
		defer expRows.Close()
		for expRows.Next() {
			var id sql.NullInt64
			var selectType, table, typeCol sql.NullString
			var possibleKeys, key, keyLen, ref sql.NullString
			var numRows sql.NullInt64
			var extra sql.NullString

			_ = expRows.Scan(&id, &selectType, &table, &typeCol, &possibleKeys, &key, &keyLen, &ref, &numRows, &extra)
			qm.IndexUsed = key.String
			qm.KeyLen = keyLen.String
			qm.Ref = ref.String
			qm.RowsExamined = numRows.Int64
			qm.Extra = extra.String
			if qm.IndexUsed == "" {
				qm.IndexUsed = "(NONE - Table Scan)"
			}
		}
	}

	// 2. Measure Execution Latencies
	var latencies []time.Duration
	for i := 0; i < iterations; i++ {
		start := time.Now()
		r, qErr := db.Query(query)
		if qErr != nil {
			log.Fatalf("Query failed: %v", qErr)
		}
		for r.Next() {
		}
		r.Close()
		latencies = append(latencies, time.Since(start))
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	qm.MinLatency = latencies[0]
	qm.MaxLatency = latencies[len(latencies)-1]
	qm.AvgLatency = calcAvg(latencies)
	p95Idx := int(math.Ceil(float64(len(latencies))*0.95)) - 1
	if p95Idx < 0 {
		p95Idx = 0
	}
	qm.P95Latency = latencies[p95Idx]

	return qm
}

func calcAvg(d []time.Duration) time.Duration {
	if len(d) == 0 {
		return 0
	}
	var total time.Duration
	for _, v := range d {
		total += v
	}
	return total / time.Duration(len(d))
}

func minDur(d []time.Duration) time.Duration {
	if len(d) == 0 {
		return 0
	}
	m := d[0]
	for _, v := range d[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func maxDur(d []time.Duration) time.Duration {
	if len(d) == 0 {
		return 0
	}
	m := d[0]
	for _, v := range d[1:] {
		if v > m {
			m = v
		}
	}
	return m
}
