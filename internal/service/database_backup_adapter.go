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
	if a.cfg != nil && a.cfg.DBType != "" {
		engine = strings.ToLower(a.cfg.DBType)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return nil, fmt.Errorf("failed creating dump output directory: %w", err)
	}

	dumpTool, _ := a.DetectCLITools()
	useCLI := (dumpTool != "" && (engine == "mariadb" || engine == "mysql"))

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
			"/*!40101 SET NAMES utf8mb4 */;\n" +
			"/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;\n" +
			"/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;\n\n"); err != nil {
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
		if _, err := writer.WriteString("/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;\n" +
			"/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;\n"); err != nil {
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

// Restore executes the SQL statements from dumpFile into the active database
func (a *DatabaseBackupAdapter) Restore(ctx context.Context, dumpFile string) error {
	if a.db == nil {
		return errors.New("database connection is nil")
	}

	if err := a.ValidateDumpFile(dumpFile); err != nil {
		return fmt.Errorf("invalid backup dump file: %w", err)
	}

	f, err := os.Open(dumpFile)
	if err != nil {
		return fmt.Errorf("failed opening dump file for restore: %w", err)
	}
	defer f.Close()

	sqlDB, err := a.db.DB()
	if err != nil {
		return fmt.Errorf("failed accessing underlying database: %w", err)
	}

	isSQLite := a.isSQLite()

	// Disable foreign keys during restore
	if isSQLite {
		_, _ = sqlDB.ExecContext(ctx, "PRAGMA foreign_keys = OFF;")
		defer func() { _, _ = sqlDB.ExecContext(context.Background(), "PRAGMA foreign_keys = ON;") }()
	} else {
		_, _ = sqlDB.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0;")
		defer func() { _, _ = sqlDB.ExecContext(context.Background(), "SET FOREIGN_KEY_CHECKS = 1;") }()
	}

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var stmtBuilder strings.Builder
	inMultiLineComment := false

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Comment and empty line handling
		if inMultiLineComment {
			if strings.Contains(line, "*/") {
				inMultiLineComment = false
			}
			continue
		}
		if strings.HasPrefix(trimmed, "/*") && !strings.HasPrefix(trimmed, "/*!") {
			if !strings.Contains(trimmed, "*/") {
				inMultiLineComment = true
			}
			continue
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		stmtBuilder.WriteString(line)
		stmtBuilder.WriteString("\n")

		if strings.HasSuffix(trimmed, ";") {
			stmt := strings.TrimSpace(stmtBuilder.String())
			stmtBuilder.Reset()

			if stmt == "" || stmt == ";" {
				continue
			}

			if _, err := sqlDB.ExecContext(ctx, stmt); err != nil {
				// Don't leak credentials or massive SQL payloads in logs
				sanitizedStmt := stmt
				if len(sanitizedStmt) > 200 {
					sanitizedStmt = sanitizedStmt[:200] + "..."
				}
				return fmt.Errorf("SQL execution failed on [%s]: %w", sanitizedStmt, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading dump file stream: %w", err)
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

	reader := bufio.NewReader(f)
	// Read first 2KB to check header
	buf := make([]byte, 2048)
	n, _ := reader.Read(buf)
	headerText := string(buf[:n])

	validHeader := strings.Contains(headerText, "--") ||
		strings.Contains(headerText, "CREATE") ||
		strings.Contains(headerText, "INSERT") ||
		strings.Contains(headerText, "/*") ||
		strings.Contains(headerText, "PRAGMA")

	if !validHeader {
		return errors.New("dump file does not contain valid SQL header or syntax markers")
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
