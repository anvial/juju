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
	var maxVersion uint64
	for v := range export.ExportVersions {
		if v > maxVersion {
			maxVersion = v
		}
	}

	tableNames, err := getTableNames(ctx, runner)
	if err != nil {
		return err
	}

	var structs []string
	var imports = make(map[string]struct{})

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
		for _, imp := range requiredImports {
			imports[imp] = struct{}{}
		}
	}

	return writeModelFile(maxVersion, structs, imports)
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

type Column struct {
	Name    string
	Type    string
	NotNull bool
}

func getTableSchema(ctx context.Context, runner *txnRunner, tableName string) ([]Column, error) {
	var columns []Column
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
			columns = append(columns, Column{
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

func generateStruct(tableName string, columns []Column) (string, []string) {
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
		sb.WriteString(fmt.Sprintf("\t%s %s `db:\"%s\"`\n", fieldName, goType, col.Name))
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

func writeModelFile(version uint64, structs []string, imports map[string]struct{}) error {
	// We should be in domain/export/generate.
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)

	// Go up one level to domain/export, then go into types/<version>.
	dir := filepath.Join(filepath.Dir(currentDir), "types", fmt.Sprintf("v%d", version))

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var out bytes.Buffer
	out.WriteString(`// Copyright 2025 Canonical Ltd. All rights reserved.
// Licensed under the AGPLv3, see LICENCE file for details.

// Code generated by go generate; DO NOT EDIT.

`)
	out.WriteString(fmt.Sprintf("package v%d\n\n", version))
	if len(imports) > 0 {
		out.WriteString("import (\n")
		// Sort imports for consistent output.
		sortedImports := make([]string, 0, len(imports))
		for imp := range imports {
			sortedImports = append(sortedImports, imp)
		}
		sort.Strings(sortedImports)
		for _, imp := range sortedImports {
			out.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
		}
		out.WriteString(")\n\n")
	}

	for _, s := range structs {
		out.WriteString(s)
		out.WriteString("\n")
	}

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		log.Printf("error formatting generated code: %v", err)
		// Write the unformatted code anyway.
		// This might be useful for diagnostics if there is an issue.
		formatted = out.Bytes()
	}

	filePath := filepath.Join(dir, "model.go")
	fmt.Printf("writing to %s\n", filePath)
	return os.WriteFile(filePath, formatted, 0644)
}
