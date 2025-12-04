// Copyright 2025 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

//go:generate go run main.go

package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"text/template"

	"github.com/canonical/sqlair"
	_ "github.com/mattn/go-sqlite3"

	"github.com/juju/juju/core/version"
	"github.com/juju/juju/domain/export"
	"github.com/juju/juju/domain/schema"
	"github.com/juju/juju/internal/database"
	"github.com/juju/juju/internal/logger"
)

// txnRunner is the simplest possible implementation of
// [core.database.TxnRunner]. It is used here to run database
// migrations and query schema metadata.
type txnRunner struct {
	db *sql.DB
}

func (r *txnRunner) Txn(ctx context.Context, f func(context.Context, *sqlair.TX) error) error {
	return database.Txn(ctx, sqlair.NewDB(r.db), f)
}

func (r *txnRunner) StdTxn(ctx context.Context, f func(context.Context, *sql.Tx) error) error {
	return database.StdTxn(ctx, r.db, f)
}

func (r *txnRunner) Dying() <-chan struct{} {
	return nil
}

func main() {
	fmt.Printf("Juju version: %s\n", version.Current)

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	runner := &txnRunner{db: db}
	m := database.NewDBMigration(runner, logger.Noop(), schema.ModelDDLForVersion(version.Current))

	ctx := context.Background()
	if err := m.Apply(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to apply migration: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Applied model schema.")

	if err := generate(ctx, runner); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate schema: %v\n", err)
		os.Exit(1)
	}
}

func generate(ctx context.Context, runner *txnRunner) error {
	var thisVersion uint64
	for v := range export.ExportVersions {
		if v > thisVersion {
			thisVersion = v
		}
	}

	tableNames, err := getTableNames(ctx, runner)
	if err != nil {
		return err
	}

	var structs, structNames, usedTableNames []string
	imports := make(map[string]struct{})

	for _, tableName := range tableNames {
		if tableName == "sqlite_sequence" {
			continue
		}
		columns, err := getTableSchema(ctx, runner, tableName)
		if err != nil {
			return err
		}

		structDef, requiredImports := generateStruct(tableName, columns)
		structs = append(structs, structDef)
		structNames = append(structNames, toCamelCase(tableName))
		usedTableNames = append(usedTableNames, tableName)
		for _, imp := range requiredImports {
			imports[imp] = struct{}{}
		}
	}

	if err := writeTypesFile(thisVersion, structs, structNames, imports); err != nil {
		return err
	}

	return writeStateModelVersionFile(thisVersion, usedTableNames, structNames)
}

func getTableNames(ctx context.Context, runner *txnRunner) ([]string, error) {
	var tableNames []string
	err := runner.StdTxn(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table'")
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return err
			}
			tableNames = append(tableNames, name)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(tableNames)
	return tableNames, nil
}

type column struct {
	Name    string
	Type    string
	NotNull bool
}

func getTableSchema(ctx context.Context, runner *txnRunner, tableName string) ([]column, error) {
	var columns []column
	query := fmt.Sprintf("PRAGMA table_info(%q)", tableName)
	err := runner.StdTxn(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var cid int
			var name, typ, dflt_value sql.NullString
			var notnull, pk int
			if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt_value, &pk); err != nil {
				return err
			}
			columns = append(columns, column{
				Name:    name.String,
				Type:    typ.String,
				NotNull: notnull != 0,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return columns, nil
}

var re = regexp.MustCompile(`_(\w)`)

func toCamelCase(s string) string {
	s = strings.ToLower(s)
	s = re.ReplaceAllStringFunc(s, func(s string) string {
		return strings.ToUpper(s[1:])
	})
	return strings.ToUpper(s[:1]) + s[1:]
}

func generateStruct(tableName string, columns []column) (string, []string) {
	structName := toCamelCase(tableName)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s struct {\n", structName))

	var imports []string

	for _, col := range columns {
		goType, imp := sqliteTypeToGoType(col.Type, col.NotNull)
		if imp != "" {
			imports = append(imports, imp)
		}
		fieldName := toCamelCase(col.Name)
		sb.WriteString(fmt.Sprintf("\t%s %s `db:%q`\n", fieldName, goType, col.Name))
	}
	sb.WriteString("}\n")
	return sb.String(), imports
}

func sqliteTypeToGoType(sqliteType string, notNull bool) (string, string) {
	var goType, imp string

	switch strings.ToUpper(sqliteType) {
	case "INTEGER", "INT":
		goType = "int64"
	case "TEXT":
		goType = "string"
	case "BOOLEAN":
		goType = "bool"
	case "DATETIME", "TIMESTAMP":
		goType = "time.Time"
		imp = "time"
	case "BLOB":
		goType = "[]byte"
	default:
		goType = "any"
	}

	if !notNull {
		goType = "*" + goType
	}
	return goType, imp
}

func writeTypesFile(version uint64, structs []string, structNames []string, imports map[string]struct{}) error {
	// We should be in domain/export/generate.
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)

	// Go up one level to domain/export, then go into types/<version>.
	dir := filepath.Join(filepath.Dir(currentDir), "types", fmt.Sprintf("v%d", version))

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Prepare import slice sorted for stable output.
	sortedImports := make([]string, 0, len(imports))
	for imp := range imports {
		sortedImports = append(sortedImports, imp)
	}
	sort.Strings(sortedImports)

	tmplPath := filepath.Join(filepath.Dir(filename), "types.tmpl")
	tmplBytes, err := os.ReadFile(tmplPath)
	if err != nil {
		return err
	}

	data := struct {
		Version     uint64
		Imports     []string
		Structs     []string
		StructNames []string
	}{
		Version:     version,
		Imports:     sortedImports,
		Structs:     structs,
		StructNames: structNames,
	}

	t := template.Must(template.New("types").Parse(string(tmplBytes)))
	var out bytes.Buffer
	if err := t.Execute(&out, data); err != nil {
		return err
	}

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		return err
	}

	filePath := filepath.Join(dir, "model.go")
	fmt.Printf("writing to %s\n", filePath)
	return os.WriteFile(filePath, formatted, 0644)
}

func writeStateModelVersionFile(version uint64, tableNames []string, structNames []string) error {
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)

	dir := filepath.Join(filepath.Dir(currentDir), "state", "model")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmplPath := filepath.Join(filepath.Dir(filename), "state.tmpl")
	tmplBytes, err := os.ReadFile(tmplPath)
	if err != nil {
		return err
	}

	data := struct {
		Version     uint64
		TableNames  []string
		StructNames []string
	}{
		Version:     version,
		TableNames:  tableNames,
		StructNames: structNames,
	}

	t := template.Must(template.New("state").Parse(string(tmplBytes)))
	var out bytes.Buffer
	if err := t.Execute(&out, data); err != nil {
		return err
	}

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		log.Printf("error formatting generated code for v%d.go: %v", version, err)
		formatted = out.Bytes()
	}

	filePath := filepath.Join(dir, fmt.Sprintf("v%d.go", version))
	fmt.Printf("writing to %s\n", filePath)
	if err := os.WriteFile(filePath, formatted, 0644); err != nil {
		return err
	}

	// Also generate a basic test that runs the ExportV<version>
	// method against the real model DB.
	testTmplPath := filepath.Join(filepath.Dir(filename), "state_test.tmpl")
	testTmplBytes, err := os.ReadFile(testTmplPath)
	if err != nil {
		return err
	}

	testData := struct {
		Version uint64
	}{
		Version: version,
	}

	testT := template.Must(template.New("state_test").Parse(string(testTmplBytes)))
	var testOut bytes.Buffer
	if err := testT.Execute(&testOut, testData); err != nil {
		return err
	}
	testFormatted, err := format.Source(testOut.Bytes())
	if err != nil {
		return err
	}

	testFilePath := filepath.Join(dir, fmt.Sprintf("v%d_test.go", version))
	fmt.Printf("writing to %s\n", testFilePath)
	if err := os.WriteFile(testFilePath, testFormatted, 0644); err != nil {
		return err
	}

	return nil
}
