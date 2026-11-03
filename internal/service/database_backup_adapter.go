package service

import (
	"bufio"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/logger"
	"datalogger/internal/model"

	"gorm.io/gorm"
)

// DatabaseBackupAdapter handles MariaDB/MySQL and SQLite database dumps and restores
type DatabaseBackupAdapter struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewDatabaseBackupAdapter(db *gorm.DB, cfg *config.Config) *DatabaseBackupAdapter {
	return &DatabaseBackupAdapter{
		db:  db,
		cfg: cfg,
	}
}

// DetectCLITools checks if mariadb-dump or mysqldump is available in system PATH
func (a *DatabaseBackupAdapter) DetectCLITools() (dumpTool string, restoreTool string) {
	// Check dump tools
	for _, tool := range []string{"mariadb-dump", "mysqldump"} {
		if path, err := exec.LookPath(tool); err == nil && path != "" {
			dumpTool = path
			break
		}
	}

	// Check restore tools
	for _, tool := range []string{"mariadb", "mysql"} {
		if path, err := exec.LookPath(tool); err == nil && path != "" {
			restoreTool = path
			break
		}
	}

	return dumpTool, restoreTool
}

// Dump executes a consistent database export to the specified outputFile
func (a *DatabaseBackupAdapter) Dump(ctx context.Context, outputFile string) (*model.DatabaseManifestInfo, error) {
	if a.db == nil {
		return nil, errors.New("database connection is nil")
	}

	start := time.Now()
	engine := "mariadb"
	if a.isSQLite() {
		engine = "sqlite"
	} else if a.cfg != nil && a.cfg.DBType != "" {
		engine = strings.ToLower(a.cfg.DBType)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return nil, fmt.Errorf("failed creating dump output directory: %w", err)
	}

	dumpTool, _ := a.DetectCLITools()
	useCLI := (!a.isSQLite() && dumpTool != "" && (engine == "mariadb" || engine == "mysql"))

	var dumpMethod string
	var err error

	if useCLI {
		dumpMethod = filepath.Base(dumpTool)
		logger.Info("Executing external database dump utility: %s", dumpMethod)
		err = a.dumpWithCLI(ctx, dumpTool, outputFile)
		if err != nil {
			logger.Warn("External %s dump failed (%v), falling back to native streaming SQL dump", dumpMethod, err)
			dumpMethod = "native-sql-stream"
			err = a.dumpWithNativeStream(ctx, outputFile)
		}
	} else {
		dumpMethod = "native-sql-stream"
		logger.Info("Executing native streaming SQL dump (engine: %s)", engine)
		err = a.dumpWithNativeStream(ctx, outputFile)
	}

	if err != nil {
		_ = os.Remove(outputFile)
		return nil, fmt.Errorf("database dump failed: %w", err)
	}

	// Validate output file and compute checksums
	fi, err := os.Stat(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed inspecting dump file: %w", err)
	}
	if fi.Size() == 0 {
		_ = os.Remove(outputFile)
		return nil, errors.New("database dump produced empty file (0 bytes)")
	}

	// Validate content header
	if err := a.ValidateDumpFile(outputFile); err != nil {
		_ = os.Remove(outputFile)
		return nil, fmt.Errorf("database dump validation failed: %w", err)
	}

	shaHex, err := a.calculateFileSHA256(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed calculating dump checksum: %w", err)
	}

	tableRows, totalRows, err := a.GetTableStats(ctx)
	if err != nil {
		logger.Warn("Failed collecting table statistics: %v", err)
	}

	dbName := "datalogger"
	dbHost := "127.0.0.1"
	if a.cfg != nil {
		if a.cfg.DBName != "" {
			dbName = a.cfg.DBName
		}
		if a.cfg.DBHost != "" {
			dbHost = a.cfg.DBHost
		}
	}

	info := &model.DatabaseManifestInfo{
		Engine:        engine,
		Host:          dbHost,
		DatabaseName:  dbName,
		DumpMethod:    dumpMethod,
		DumpFileName:  filepath.Base(outputFile),
		DumpSHA256:    shaHex,
		DumpSizeBytes: fi.Size(),
		TableCount:    len(tableRows),
		TotalRows:     totalRows,
		TableRows:     tableRows,
	}

	logger.Info("Database dump completed in %v (method: %s, size: %d bytes, tables: %d, rows: %d)",
		time.Since(start), dumpMethod, fi.Size(), len(tableRows), totalRows)

	return info, nil
}

// dumpWithCLI executes mariadb-dump or mysqldump with single transaction consistency
func (a *DatabaseBackupAdapter) dumpWithCLI(ctx context.Context, tool string, outputFile string) error {
	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	host := "127.0.0.1"
	port := "3306"
	user := "root"
	pass := ""
	dbName := "datalogger"

	if a.cfg != nil {
		if a.cfg.DBHost != "" {
			host = a.cfg.DBHost
		}
		if a.cfg.DBPort != "" {
			port = a.cfg.DBPort
		}
		if a.cfg.DBUser != "" {
			user = a.cfg.DBUser
		}
		pass = a.cfg.DBPassword
		if a.cfg.DBName != "" {
			dbName = a.cfg.DBName
		}
	}

	args := []string{
		fmt.Sprintf("--host=%s", host),
		fmt.Sprintf("--port=%s", port),
		fmt.Sprintf("--user=%s", user),
		"--single-transaction",
		"--quick",
		"--skip-lock-tables",
		"--routines",
		"--triggers",
		fmt.Sprintf("--ignore-table=%s.backup_records", dbName),
		"--default-character-set=utf8mb4",
		"--hex-blob",
	}

	if pass != "" {
		args = append(args, fmt.Sprintf("--password=%s", pass))
	}
	args = append(args, dbName)

	cmd := exec.CommandContext(ctx, tool, args...)
	cmd.Stdout = outFile

	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	// Redact password in debug logs
	redactedArgs := make([]string, len(args))
	for i, arg := range args {
		if strings.HasPrefix(arg, "--password=") {
			redactedArgs[i] = "--password=***"
		} else {
			redactedArgs[i] = arg
		}
	}
	logger.Debug("Invoking database dump process: %s %s", tool, strings.Join(redactedArgs, " "))

	if err := cmd.Run(); err != nil {
		errMsg := a.redactSecrets(stderrBuf.String())
		return fmt.Errorf("process error (exit code %d): %s: %w", cmd.ProcessState.ExitCode(), errMsg, err)
	}

	return nil
}

// dumpWithNativeStream creates a pure-Go streaming logical SQL dump via database/sql
func (a *DatabaseBackupAdapter) dumpWithNativeStream(ctx context.Context, outputFile string) error {
	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := bufio.NewWriterSize(outFile, 64*1024)
	defer writer.Flush()

	// Write dump metadata header
	engine := "mariadb"
	if a.cfg != nil && a.cfg.DBType != "" {
		engine = a.cfg.DBType
	}
	isSQLite := a.isSQLite()
	if isSQLite {
		engine = "sqlite"
	}

	header := fmt.Sprintf("-- ========================================================\n"+
		"-- Datalogger Analysis Application Database Logical Dump\n"+
		"-- Generated: %s UTC\n"+
		"-- Engine: %s\n"+
		"-- Format: SQL Dump v1.0\n"+
		"-- ========================================================\n\n",
		time.Now().UTC().Format(time.RFC3339), engine)
	if _, err := writer.WriteString(header); err != nil {
		return err
	}

	if isSQLite {
		if _, err := writer.WriteString("PRAGMA foreign_keys = OFF;\nBEGIN TRANSACTION;\n\n"); err != nil {
			return err
		}
	} else {
		if _, err := writer.WriteString("/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;\n" +
			"/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;\n" +
			"/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;\n" +
			"/*!40101 SET NAMES utf8mb4 */;\n" +
			"/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;\n" +
			"/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;\n" +
			"/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;\n" +
			"/*!40101 SET @OLD_AUTOCOMMIT=@@AUTOCOMMIT, AUTOCOMMIT=0 */;\n\n"); err != nil {
			return err
		}
	}

	// Retrieve table names
	tables, err := a.getTableList(ctx, isSQLite)
	if err != nil {
		return fmt.Errorf("failed retrieving table list: %w", err)
	}

	sqlDB, err := a.db.DB()
	if err != nil {
		return fmt.Errorf("failed accessing underlying sql.DB: %w", err)
	}

	for _, table := range tables {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 1. Table structure
		if _, err := writer.WriteString(fmt.Sprintf("--\n-- Table structure for table `%s`\n--\n\n", table)); err != nil {
			return err
		}
		if _, err := writer.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n", table)); err != nil {
			return err
		}

		ddl, err := a.getTableDDL(ctx, table, isSQLite)
		if err != nil {
			logger.Warn("Failed retrieving DDL for table %s: %v", table, err)
			continue
		}
		if !strings.HasSuffix(strings.TrimSpace(ddl), ";") {
			ddl = strings.TrimSpace(ddl) + ";"
		}
		if _, err := writer.WriteString(ddl + "\n\n"); err != nil {
			return err
		}

		// 2. Table data
		if _, err := writer.WriteString(fmt.Sprintf("--\n-- Dumping data for table `%s`\n--\n\n", table)); err != nil {
			return err
		}

		if err := a.dumpTableData(ctx, sqlDB, table, writer, isSQLite); err != nil {
			return fmt.Errorf("failed dumping data for table %s: %w", table, err)
		}
		if _, err := writer.WriteString("\n"); err != nil {
			return err
		}
	}

	// Trailing footer
	if isSQLite {
		if _, err := writer.WriteString("COMMIT;\nPRAGMA foreign_keys = ON;\n"); err != nil {
			return err
		}
	} else {
		if _, err := writer.WriteString("COMMIT;\n" +
			"/*!40101 SET AUTOCOMMIT=IFNULL(@OLD_AUTOCOMMIT, 1) */;\n" +
			"/*!40014 SET UNIQUE_CHECKS=IFNULL(@OLD_UNIQUE_CHECKS, 1) */;\n" +
			"/*!40014 SET FOREIGN_KEY_CHECKS=IFNULL(@OLD_FOREIGN_KEY_CHECKS, 1) */;\n" +
			"/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;\n" +
			"/*!40101 SET CHARACTER_SET_CLIENT=IFNULL(@OLD_CHARACTER_SET_CLIENT, 'utf8mb4') */;\n" +
			"/*!40101 SET CHARACTER_SET_RESULTS=IFNULL(@OLD_CHARACTER_SET_RESULTS, 'utf8mb4') */;\n" +
			"/*!40101 SET COLLATION_CONNECTION=IFNULL(@OLD_COLLATION_CONNECTION, 'utf8mb4_general_ci') */;\n"); err != nil {
			return err
		}
	}

	completionFooter := fmt.Sprintf("\n-- Dump completed on %s UTC\n", time.Now().UTC().Format(time.RFC3339))
	_, err = writer.WriteString(completionFooter)
	return err
}

func (a *DatabaseBackupAdapter) dumpTableData(ctx context.Context, sqlDB *sql.DB, table string, writer *bufio.Writer, isSQLite bool) error {
	query := fmt.Sprintf("SELECT * FROM `%s`", table)
	rows, err := sqlDB.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	colCount := len(cols)

	quotedCols := make([]string, colCount)
	for i, c := range cols {
		quotedCols[i] = fmt.Sprintf("`%s`", c)
	}
	colHeader := strings.Join(quotedCols, ", ")

	batchValues := make([]string, 0, 50)
	rowCount := 0

	for rows.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		values := make([]interface{}, colCount)
		valuePtrs := make([]interface{}, colCount)
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return err
		}

		rowStrs := make([]string, colCount)
		for i, val := range values {
			rowStrs[i] = a.formatSQLValue(val)
		}

		batchValues = append(batchValues, "("+strings.Join(rowStrs, ", ")+")")
		rowCount++

		if len(batchValues) >= 50 {
			stmt := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES\n%s;\n",
				table, colHeader, strings.Join(batchValues, ",\n"))
			if _, err := writer.WriteString(stmt); err != nil {
				return err
			}
			batchValues = batchValues[:0]
		}
	}

	if len(batchValues) > 0 {
		stmt := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES\n%s;\n",
			table, colHeader, strings.Join(batchValues, ",\n"))
		if _, err := writer.WriteString(stmt); err != nil {
			return err
		}
	}

	return rows.Err()
}

func (a *DatabaseBackupAdapter) formatSQLValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}
	switch v := val.(type) {
	case []byte:
		str := string(v)
		return "'" + strings.ReplaceAll(strings.ReplaceAll(str, "\\", "\\\\"), "'", "''") + "'"
	case string:
		return "'" + strings.ReplaceAll(strings.ReplaceAll(v, "\\", "\\\\"), "'", "''") + "'"
	case time.Time:
		return "'" + v.UTC().Format("2006-01-02 15:04:05.999999") + "'"
	case bool:
		if v {
			return "1"
		}
		return "0"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	default:
		return "'" + strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''") + "'"
	}
}

func (a *DatabaseBackupAdapter) getTableList(ctx context.Context, isSQLite bool) ([]string, error) {
	sqlDB, err := a.db.DB()
	if err != nil {
		return nil, err
	}

	var query string
	if isSQLite {
		query = "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name"
	} else {
		query = "SHOW FULL TABLES WHERE Table_type = 'BASE TABLE'"
	}

	rows, err := sqlDB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		var tableType string
		if isSQLite {
			if err := rows.Scan(&name); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(&name, &tableType); err != nil {
				return nil, err
			}
		}
		if strings.EqualFold(name, "backup_records") {
			continue
		}
		tables = append(tables, name)
	}

	return tables, rows.Err()
}

func (a *DatabaseBackupAdapter) getTableDDL(ctx context.Context, table string, isSQLite bool) (string, error) {
	sqlDB, err := a.db.DB()
	if err != nil {
		return "", err
	}

	if isSQLite {
		var sqlStr string
		err := sqlDB.QueryRowContext(ctx, "SELECT sql FROM sqlite_master WHERE type='table' AND name = ?", table).Scan(&sqlStr)
		return sqlStr, err
	}

	var tableName string
	var ddl string
	err = sqlDB.QueryRowContext(ctx, fmt.Sprintf("SHOW CREATE TABLE `%s`", table)).Scan(&tableName, &ddl)
	return ddl, err
}

var (
	reRestoreAutocommit       = regexp.MustCompile(`(^|[^@a-zA-Z0-9_]|@@)(?i:autocommit)\s*=\s*(?i:@old_autocommit)\b`)
	reRestoreForeignKeyChecks = regexp.MustCompile(`(^|[^@a-zA-Z0-9_]|@@)(?i:foreign_key_checks)\s*=\s*(?i:@old_foreign_key_checks)\b`)
	reRestoreUniqueChecks     = regexp.MustCompile(`(^|[^@a-zA-Z0-9_]|@@)(?i:unique_checks)\s*=\s*(?i:@old_unique_checks)\b`)
	reRestoreSQLNotes         = regexp.MustCompile(`(^|[^@a-zA-Z0-9_]|@@)(?i:sql_notes)\s*=\s*(?i:@old_sql_notes)\b`)
	reRestoreSQLMode          = regexp.MustCompile(`(^|[^@a-zA-Z0-9_]|@@)(?i:sql_mode)\s*=\s*(?i:@old_sql_mode)\b`)
	reRestoreTimeZone         = regexp.MustCompile(`(^|[^@a-zA-Z0-9_]|@@)(?i:time_zone)\s*=\s*(?i:@old_time_zone)\b`)
	reRestoreCharSetClient    = regexp.MustCompile(`(^|[^@a-zA-Z0-9_]|@@)(?i:character_set_client)\s*=\s*(?i:@old_character_set_client)\b`)
	reRestoreCharSetResults   = regexp.MustCompile(`(^|[^@a-zA-Z0-9_]|@@)(?i:character_set_results)\s*=\s*(?i:@old_character_set_results)\b`)
	reRestoreCollationConn    = regexp.MustCompile(`(^|[^@a-zA-Z0-9_]|@@)(?i:collation_connection)\s*=\s*(?i:@old_collation_connection)\b`)
	reDefinerComment          = regexp.MustCompile(`/\*![0-9]+\s+DEFINER\s*=\s*[^*]+\*/`)
	reDefinerPlain            = regexp.MustCompile(`(?i)\bDEFINER\s*=\s*(?:` + "`[^`]+`" + `|'[^']+'|"[^"]+"|[a-zA-Z0-9_]+)\s*@\s*(?:` + "`[^`]+`" + `|'[^']+'|"[^"]+"|[a-zA-Z0-9_.%-]+)`)
)

const mysqlRestoreSessionPreamble = "SET @OLD_AUTOCOMMIT=@@AUTOCOMMIT;\n" +
	"SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT;\n" +
	"SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS;\n" +
	"SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION;\n" +
	"SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;\n" +
	"SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;\n" +
	"SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO';\n" +
	"SET @OLD_TIME_ZONE=@@TIME_ZONE;\n" +
	"SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0;\n"

func sanitizeMySQLSessionRestore(stmt string) string {
	if len(stmt) > 1024 || !strings.Contains(strings.ToUpper(stmt), "@OLD_") {
		return stmt
	}
	stmt = reRestoreAutocommit.ReplaceAllString(stmt, "${1}AUTOCOMMIT=IFNULL(@OLD_AUTOCOMMIT, 1)")
	stmt = reRestoreForeignKeyChecks.ReplaceAllString(stmt, "${1}FOREIGN_KEY_CHECKS=IFNULL(@OLD_FOREIGN_KEY_CHECKS, 1)")
	stmt = reRestoreUniqueChecks.ReplaceAllString(stmt, "${1}UNIQUE_CHECKS=IFNULL(@OLD_UNIQUE_CHECKS, 1)")
	stmt = reRestoreSQLNotes.ReplaceAllString(stmt, "${1}SQL_NOTES=IFNULL(@OLD_SQL_NOTES, 1)")
	stmt = reRestoreSQLMode.ReplaceAllString(stmt, "${1}SQL_MODE=IFNULL(@OLD_SQL_MODE, '')")
	stmt = reRestoreTimeZone.ReplaceAllString(stmt, "${1}TIME_ZONE=IFNULL(@OLD_TIME_ZONE, 'SYSTEM')")
	stmt = reRestoreCharSetClient.ReplaceAllString(stmt, "${1}CHARACTER_SET_CLIENT=IFNULL(@OLD_CHARACTER_SET_CLIENT, 'utf8mb4')")
	stmt = reRestoreCharSetResults.ReplaceAllString(stmt, "${1}CHARACTER_SET_RESULTS=IFNULL(@OLD_CHARACTER_SET_RESULTS, 'utf8mb4')")
	stmt = reRestoreCollationConn.ReplaceAllString(stmt, "${1}COLLATION_CONNECTION=IFNULL(@OLD_COLLATION_CONNECTION, 'utf8mb4_general_ci')")
	return stmt
}

// Restore executes the SQL statements from dumpFile into the active database
func (a *DatabaseBackupAdapter) Restore(ctx context.Context, dumpFile string) error {
	return a.RestoreWithProgress(ctx, dumpFile, nil)
}

// RestoreWithProgress executes the SQL statements from dumpFile into the active database and reports [0.0..1.0] progress
func (a *DatabaseBackupAdapter) RestoreWithProgress(ctx context.Context, dumpFile string, onProgress func(fraction float64)) error {
	if a.db == nil {
		return errors.New("database connection is nil")
	}

	if err := a.ValidateDumpFile(dumpFile); err != nil {
		return fmt.Errorf("invalid backup dump file: %w", err)
	}

	isSQLite := a.isSQLite()
	engine := "mariadb"
	if isSQLite {
		engine = "sqlite"
	} else if a.cfg != nil && a.cfg.DBType != "" {
		engine = strings.ToLower(a.cfg.DBType)
	}

	_, restoreTool := a.DetectCLITools()
	useCLI := (!isSQLite && restoreTool != "" && (engine == "mariadb" || engine == "mysql"))

	if useCLI {
		logger.Info("Executing external database restore utility: %s", filepath.Base(restoreTool))
		if err := a.restoreWithCLI(ctx, restoreTool, dumpFile, onProgress); err == nil {
			return nil
		} else {
			logger.Warn("External %s restore failed (%v), falling back to native pinned-connection SQL restore", filepath.Base(restoreTool), err)
		}
	}

	return a.restoreWithNativeStream(ctx, dumpFile, isSQLite, onProgress)
}

func (a *DatabaseBackupAdapter) restoreWithCLI(ctx context.Context, tool string, dumpFile string, onProgress func(fraction float64)) error {
	fi, err := os.Stat(dumpFile)
	if err != nil {
		return err
	}
	totalSize := fi.Size()

	f, err := os.Open(dumpFile)
	if err != nil {
		return err
	}
	defer f.Close()

	host := "127.0.0.1"
	port := "3306"
	user := "root"
	pass := ""
	dbName := "datalogger"

	if a.cfg != nil {
		if a.cfg.DBHost != "" {
			host = a.cfg.DBHost
		}
		if a.cfg.DBPort != "" {
			port = a.cfg.DBPort
		}
		if a.cfg.DBUser != "" {
			user = a.cfg.DBUser
		}
		pass = a.cfg.DBPassword
		if a.cfg.DBName != "" {
			dbName = a.cfg.DBName
		}
	}

	args := []string{
		fmt.Sprintf("--host=%s", host),
		fmt.Sprintf("--port=%s", port),
		fmt.Sprintf("--user=%s", user),
		"--default-character-set=utf8mb4",
	}
	if pass != "" {
		args = append(args, fmt.Sprintf("--password=%s", pass))
	}
	args = append(args, dbName)

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		bw := bufio.NewWriterSize(pw, 256*1024)
		if _, wErr := bw.WriteString(mysqlRestoreSessionPreamble); wErr != nil {
			_ = pw.CloseWithError(wErr)
			return
		}

		br := bufio.NewReaderSize(f, 256*1024)
		var bytesRead int64
		var lastReport int64

		for {
			line, rErr := br.ReadString('\n')
			if len(line) > 0 {
				bytesRead += int64(len(line))
				if onProgress != nil && totalSize > 0 && (bytesRead-lastReport >= 512*1024 || bytesRead == totalSize) {
					lastReport = bytesRead
					onProgress(float64(bytesRead) / float64(totalSize))
				}

				trimmed := strings.TrimSpace(line)
				// Skip MariaDB 10.11+/11.x CLI sandbox directive that breaks cross-version clients
				if strings.HasPrefix(trimmed, "/*M!999999") {
					if rErr == io.EOF {
						break
					}
					continue
				}
				if len(trimmed) < 1024 && strings.Contains(strings.ToUpper(trimmed), "@OLD_") {
					line = sanitizeMySQLSessionRestore(line)
				}
				if _, wErr := bw.WriteString(line); wErr != nil {
					_ = pw.CloseWithError(wErr)
					return
				}
			}
			if rErr == io.EOF {
				break
			}
			if rErr != nil {
				_ = pw.CloseWithError(rErr)
				return
			}
		}

		_, _ = bw.WriteString("\nCOMMIT;\nSET FOREIGN_KEY_CHECKS=1;\nSET UNIQUE_CHECKS=1;\nSET AUTOCOMMIT=1;\n")
		_ = bw.Flush()
	}()

	cmd := exec.CommandContext(ctx, tool, args...)
	cmd.Stdin = pr

	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		_ = pr.Close()
		errMsg := a.redactSecrets(strings.TrimSpace(stderrBuf.String()))
		return fmt.Errorf("CLI restore error: %s: %w", errMsg, err)
	}

	return nil
}

func (a *DatabaseBackupAdapter) restoreWithNativeStream(ctx context.Context, dumpFile string, isSQLite bool, onProgress func(fraction float64)) error {
	fi, err := os.Stat(dumpFile)
	if err != nil {
		return fmt.Errorf("failed statting dump file for restore: %w", err)
	}
	totalSize := fi.Size()

	f, err := os.Open(dumpFile)
	if err != nil {
		return fmt.Errorf("failed opening dump file for restore: %w", err)
	}
	defer f.Close()

	sqlDB, err := a.db.DB()
	if err != nil {
		return fmt.Errorf("failed accessing underlying database: %w", err)
	}

	// Pin a SINGLE dedicated database connection for the entire restore so session variables
	// (@OLD_AUTOCOMMIT, @OLD_FOREIGN_KEY_CHECKS, @OLD_SQL_MODE), AUTOCOMMIT=0 transactions,
	// and table locks are NEVER split across pooled connections.
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed acquiring dedicated database connection for restore: %w", err)
	}
	defer func() {
		cleanCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if isSQLite {
			_, _ = conn.ExecContext(cleanCtx, "PRAGMA foreign_keys = ON;")
		} else {
			_, _ = conn.ExecContext(cleanCtx, "COMMIT")
			_, _ = conn.ExecContext(cleanCtx, "UNLOCK TABLES")
			_, _ = conn.ExecContext(cleanCtx, "SET AUTOCOMMIT = 1")
			_, _ = conn.ExecContext(cleanCtx, "SET FOREIGN_KEY_CHECKS = 1")
			_, _ = conn.ExecContext(cleanCtx, "SET UNIQUE_CHECKS = 1")
		}
		_ = conn.Close()
	}()

	if isSQLite {
		_, _ = conn.ExecContext(ctx, "PRAGMA foreign_keys = OFF;")
	} else {
		for _, initStmt := range []string{
			"SET @OLD_AUTOCOMMIT=@@AUTOCOMMIT",
			"SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT",
			"SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS",
			"SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION",
			"SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0",
			"SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0",
			"SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO'",
			"SET @OLD_TIME_ZONE=@@TIME_ZONE",
			"SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0",
		} {
			_, _ = conn.ExecContext(ctx, initStmt)
		}
	}

	var lastReport int64
	progressCb := func(bytesRead int64) {
		if onProgress != nil && totalSize > 0 && (bytesRead-lastReport >= 512*1024 || bytesRead >= totalSize) {
			lastReport = bytesRead
			frac := float64(bytesRead) / float64(totalSize)
			if frac > 1.0 {
				frac = 1.0
			}
			onProgress(frac)
		}
	}

	return streamSQLStatements(ctx, f, !isSQLite, progressCb, func(rawStmt string) error {
		stmt := rawStmt
		if !isSQLite {
			stmt = sanitizeMySQLSessionRestore(stmt)
		}

		_, execErr := conn.ExecContext(ctx, stmt)
		if execErr == nil {
			return nil
		}

		errLower := strings.ToLower(execErr.Error())

		// 1. Ignore empty query result when versioned conditional comment (e.g. /*M!110400 ... */) evaluates to no-op on older server
		if strings.Contains(errLower, "query was empty") || strings.Contains(errLower, "error 1065") {
			return nil
		}

		// 2. Non-critical session variable restore tolerance (e.g. MariaDB-only @OLD_NOTE_VERBOSITY on MySQL or uninitialized @OLD_* variable)
		upperStmt := strings.ToUpper(strings.TrimSpace(stmt))
		isSetOrCommentSet := strings.HasPrefix(upperStmt, "SET ") ||
			(strings.HasPrefix(upperStmt, "/*") && strings.Contains(upperStmt, "SET ") && len(stmt) < 512)
		if !isSQLite && isSetOrCommentSet && (strings.Contains(errLower, "error 1231") ||
			strings.Contains(errLower, "can't be set to the value of 'null'") ||
			strings.Contains(errLower, "error 1193") ||
			strings.Contains(errLower, "unknown system variable")) {
			logger.Warn("Skipping non-critical session variable statement during restore [%s]: %v", truncateSQLForLog(stmt), execErr)
			return nil
		}

		// 3. Cross-version MariaDB 11.4+/11.8 -> MariaDB 10.x / MySQL 8.x collation compatibility fallback (uca1400 -> unicode_ci)
		if !isSQLite && strings.Contains(errLower, "unknown collation") && strings.Contains(stmt, "uca1400") {
			compatStmt := strings.ReplaceAll(stmt, "utf8mb4_uca1400_ai_ci", "utf8mb4_unicode_ci")
			compatStmt = strings.ReplaceAll(compatStmt, "utf8mb3_uca1400_ai_ci", "utf8mb3_unicode_ci")
			compatStmt = strings.ReplaceAll(compatStmt, "_uca1400_ai_ci", "_unicode_ci")
			compatStmt = strings.ReplaceAll(compatStmt, "_uca1400_as_cs", "_bin")
			if _, retryErr := conn.ExecContext(ctx, compatStmt); retryErr == nil {
				return nil
			}
		}

		// 4. Cross-host DEFINER privilege fallback for views/triggers/routines when restoring under non-SUPER DB user
		if !isSQLite && (strings.Contains(errLower, "error 1227") || strings.Contains(errLower, "super") || strings.Contains(errLower, "set user")) && strings.Contains(upperStmt, "DEFINER") {
			noDefStmt := reDefinerComment.ReplaceAllString(stmt, "")
			noDefStmt = reDefinerPlain.ReplaceAllString(noDefStmt, "")
			if _, retryErr := conn.ExecContext(ctx, noDefStmt); retryErr == nil {
				return nil
			}
		}

		return fmt.Errorf("SQL execution failed on [%s]: %w", truncateSQLForLog(stmt), execErr)
	})
}

func truncateSQLForLog(stmt string) string {
	if len(stmt) > 200 {
		return stmt[:200] + "..."
	}
	return stmt
}

// streamSQLStatements reads arbitrary-length SQL statements from r without bufio.Scanner's 10MB token limit,
// tracking string quotes, DELIMITER changes, MySQL /*! ... */ comments, and MariaDB /*M! ... */ comments.
func streamSQLStatements(
	ctx context.Context,
	r io.Reader,
	backslashEscapes bool,
	onBytesRead func(int64),
	execFn func(stmt string) error,
) error {
	reader := bufio.NewReaderSize(r, 256*1024)

	var stmtBuf strings.Builder
	delimiter := ";"
	inBlockComment := false
	inExecComment := false
	var quote byte
	var bytesProcessed int64

	processLine := func(line string) error {
		if !inBlockComment && !inExecComment && quote == 0 {
			trimmed := strings.TrimSpace(line)
			if stmtBuf.Len() == 0 {
				if trimmed == "" || strings.HasPrefix(trimmed, "--") || (backslashEscapes && strings.HasPrefix(trimmed, "#")) {
					return nil
				}
				if strings.HasPrefix(trimmed, "/*M!999999") {
					if !strings.Contains(trimmed, "*/") {
						inBlockComment = true
					}
					return nil
				}
			}
			if len(trimmed) >= 9 && strings.EqualFold(trimmed[:9], "DELIMITER") &&
				(len(trimmed) == 9 || trimmed[9] == ' ' || trimmed[9] == '\t') &&
				strings.TrimSpace(stmtBuf.String()) == "" {
				stmtBuf.Reset()
				newDelim := strings.TrimSpace(trimmed[9:])
				if newDelim != "" {
					delimiter = newDelim
				}
				return nil
			}
		}

		for i := 0; i < len(line); i++ {
			if inBlockComment {
				if line[i] == '*' && i+1 < len(line) && line[i+1] == '/' {
					inBlockComment = false
					i++
					if stmtBuf.Len() > 0 {
						stmtBuf.WriteByte(' ')
					}
				}
				continue
			}

			if quote != 0 {
				ch := line[i]
				stmtBuf.WriteByte(ch)
				if backslashEscapes && ch == '\\' && (quote == '\'' || quote == '"') {
					if i+1 < len(line) {
						i++
						stmtBuf.WriteByte(line[i])
					}
					continue
				}
				if ch == quote {
					if i+1 < len(line) && line[i+1] == quote {
						i++
						stmtBuf.WriteByte(line[i])
						continue
					}
					quote = 0
				}
				continue
			}

			if inExecComment {
				if line[i] == '*' && i+1 < len(line) && line[i+1] == '/' {
					stmtBuf.WriteString("*/")
					inExecComment = false
					i++
					continue
				}
				if line[i] == '\'' || line[i] == '"' || line[i] == '`' {
					quote = line[i]
					stmtBuf.WriteByte(line[i])
					continue
				}
				if line[i] == '-' && i+1 < len(line) && line[i+1] == '-' &&
					(i+2 == len(line) || line[i+2] == ' ' || line[i+2] == '\t' || line[i+2] == '\r' || line[i+2] == '\n') &&
					!strings.Contains(line[i+2:], "*/") {
					stmtBuf.WriteByte('\n')
					break
				}
				stmtBuf.WriteByte(line[i])
				continue
			}

			// Normal top-level SQL state
			dLen := len(delimiter)
			if dLen > 0 && i+dLen <= len(line) && line[i:i+dLen] == delimiter {
				stmt := strings.TrimSpace(stmtBuf.String())
				stmtBuf.Reset()
				if stmt != "" {
					if err := execFn(stmt); err != nil {
						return err
					}
				}
				i += dLen - 1
				continue
			}

			if line[i] == '-' && i+1 < len(line) && line[i+1] == '-' &&
				(i+2 == len(line) || line[i+2] == ' ' || line[i+2] == '\t' || line[i+2] == '\r' || line[i+2] == '\n') {
				if stmtBuf.Len() > 0 {
					stmtBuf.WriteByte('\n')
				}
				break
			}

			if backslashEscapes && line[i] == '#' {
				if stmtBuf.Len() > 0 {
					stmtBuf.WriteByte('\n')
				}
				break
			}

			if line[i] == '/' && i+1 < len(line) && line[i+1] == '*' {
				rest := line[i:]
				if strings.HasPrefix(rest, "/*M!999999") {
					if endIdx := strings.Index(rest[2:], "*/"); endIdx >= 0 {
						i += 2 + endIdx + 1
					} else {
						inBlockComment = true
						break
					}
					continue
				}
				if strings.HasPrefix(rest, "/*!") || strings.HasPrefix(rest, "/*M!") {
					inExecComment = true
					if strings.HasPrefix(rest, "/*M!") {
						stmtBuf.WriteString("/*M!")
						i += 3
					} else {
						stmtBuf.WriteString("/*!")
						i += 2
					}
					continue
				}
				inBlockComment = true
				i++
				continue
			}

			if line[i] == '\'' || line[i] == '"' || line[i] == '`' {
				quote = line[i]
				stmtBuf.WriteByte(line[i])
				continue
			}

			if stmtBuf.Len() == 0 && (line[i] == ' ' || line[i] == '\t' || line[i] == '\r' || line[i] == '\n') {
				continue
			}
			stmtBuf.WriteByte(line[i])
		}

		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, readErr := reader.ReadString('\n')
		if len(line) > 0 {
			bytesProcessed += int64(len(line))
			if onBytesRead != nil {
				onBytesRead(bytesProcessed)
			}
			if err := processLine(line); err != nil {
				return err
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("error reading dump file stream: %w", readErr)
		}
	}

	if rem := strings.TrimSpace(stmtBuf.String()); rem != "" {
		if err := execFn(rem); err != nil {
			return err
		}
	}

	return nil
}

// GetTableStats returns row counts per table and total rows in the database
func (a *DatabaseBackupAdapter) GetTableStats(ctx context.Context) (map[string]int64, int64, error) {
	if a.db == nil {
		return nil, 0, errors.New("database connection is nil")
	}

	isSQLite := a.isSQLite()

	tables, err := a.getTableList(ctx, isSQLite)
	if err != nil {
		return nil, 0, err
	}

	sqlDB, err := a.db.DB()
	if err != nil {
		return nil, 0, err
	}

	stats := make(map[string]int64)
	var totalRows int64

	for _, t := range tables {
		var count int64
		row := sqlDB.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM `%s`", t))
		if err := row.Scan(&count); err == nil {
			stats[t] = count
			totalRows += count
		}
	}

	return stats, totalRows, nil
}

// ValidateDumpFile verifies that dumpFile exists, is non-empty, and contains valid SQL syntax markers
func (a *DatabaseBackupAdapter) ValidateDumpFile(dumpFile string) error {
	fi, err := os.Stat(dumpFile)
	if err != nil {
		return err
	}
	if fi.Size() == 0 {
		return errors.New("dump file is empty (0 bytes)")
	}

	f, err := os.Open(dumpFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read first 2KB to check header
	buf := make([]byte, 2048)
	n, _ := f.Read(buf)
	headerText := string(buf[:n])

	validHeader := strings.Contains(headerText, "--") ||
		strings.Contains(headerText, "CREATE") ||
		strings.Contains(headerText, "INSERT") ||
		strings.Contains(headerText, "/*") ||
		strings.Contains(headerText, "PRAGMA")

	if !validHeader {
		return errors.New("dump file does not contain valid SQL header or syntax markers")
	}

	// Verify tail of file for valid SQL completion or termination
	if fi.Size() > 64 {
		seekOffset := fi.Size() - 2048
		if seekOffset < 0 {
			seekOffset = 0
		}
		if _, err := f.Seek(seekOffset, io.SeekStart); err == nil {
			tailBuf := make([]byte, 2048)
			nTail, _ := f.Read(tailBuf)
			tailText := string(tailBuf[:nTail])
			trimmedTail := strings.TrimSpace(tailText)
			// Must end with a comment or semicolon or recognized footer
			validTail := strings.HasSuffix(trimmedTail, ";") ||
				strings.Contains(tailText, "-- Dump completed") ||
				strings.Contains(tailText, "COMMIT") ||
				strings.Contains(tailText, "/*!40101") ||
				strings.Contains(tailText, "/*!40014") ||
				strings.Contains(tailText, "/*M!") ||
				strings.HasSuffix(trimmedTail, "--")
			if !validTail {
				return errors.New("dump file appears incomplete or truncated (missing SQL termination)")
			}
		}
	}

	return nil
}

func (a *DatabaseBackupAdapter) calculateFileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (a *DatabaseBackupAdapter) redactSecrets(input string) string {
	passRegex := regexp.MustCompile(`(?i)(password|pass|pwd)\s*=\s*['"]?[^'"\s]+['"]?`)
	return passRegex.ReplaceAllString(input, "${1}=***")
}

func (a *DatabaseBackupAdapter) isSQLite() bool {
	if a.db != nil && a.db.Dialector != nil && a.db.Dialector.Name() == "sqlite" {
		return true
	}
	if a.cfg != nil && strings.Contains(strings.ToLower(a.cfg.DBType), "sqlite") {
		return true
	}
	return false
}
