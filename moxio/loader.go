package moxio

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"github.com/jackc/pgx/v4"
	"github.com/lithictech/moxpopuli/internal"
	"github.com/pkg/errors"
	"io"
	"net/url"
	"os"
	"strings"
)

func LoadIterator(ctx context.Context, uri, arg string) (Iterator, error) {
	pl, err := NewLoader(ctx, uri)
	if err != nil {
		return nil, err
	}
	return pl.Iterator(ctx, arg)
}

func LoadOne(ctx context.Context, uri, arg string) (interface{}, error) {
	iter, err := LoadIterator(ctx, uri, arg)
	if err != nil {
		return nil, err
	}
	if !iter.Next() {
		return nil, nil
	}
	o, err := iter.Read(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "reading single loader item")
	}
	return o, nil
}

func LoadOneMap(ctx context.Context, uri, arg string) (map[string]interface{}, error) {
	o, err := LoadOne(ctx, uri, arg)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return make(map[string]interface{}, 4), nil
	}
	return o.(map[string]interface{}), nil
}

type Loader interface {
	Iterator(ctx context.Context, arg string) (Iterator, error)
}

type Iterator interface {
	// Next prepares the next row for reading. It returns true if there is another
	// row and false if no more rows are available.
	Next() bool
	// Read reads the given object/data with the row value.
	// It is an error to call Read without
	// first calling Next() and checking that it returned true.
	Read(context.Context) (interface{}, error)
	Close() error
}

func NewLoader(ctx context.Context, uri string) (Loader, error) {
	if uri == "." || uri == "" {
		return noopLoader{}, nil
	}
	if uri == "-" {
		return &jsonLineReader{R: os.Stdin}, nil
	}
	if uri == "_" {
		return &jsonVerbatimLoader{}, nil
	}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "postgres" {
		return (&postgresLoader{url: uri}).Connect(ctx)
	}
	if u.Scheme == "file" {
		return fileLoader{path: internal.FileUriPath(u)}, nil
	}
	return nil, errors.Errorf("unknown loader for uri '%s'", uri)
}

// Load data from a PG query. The query must select a single row that can be parsed to JSON.
type postgresLoader struct {
	url  string
	conn *pgx.Conn
}

func (m *postgresLoader) Connect(ctx context.Context) (*postgresLoader, error) {
	conn, err := pgx.Connect(ctx, m.url)
	if err != nil {
		return nil, err
	}
	m.conn = conn
	return m, nil
}

func (m *postgresLoader) Iterator(ctx context.Context, sql string) (Iterator, error) {
	return (&postgresIterator{}).query(ctx, m.conn, sql)
}

type postgresIterator struct {
	rows pgx.Rows
}

func (m *postgresIterator) query(ctx context.Context, conn *pgx.Conn, query string) (Iterator, error) {
	var err error
	m.rows, err = conn.Query(ctx, query)
	return m, err
}

func (m *postgresIterator) Next() bool {
	return m.rows.Next()
}

func (m *postgresIterator) Read(_ context.Context) (interface{}, error) {
	values, err := m.rows.Values()
	if err != nil {
		return nil, err
	}
	if len(values) == 1 {
		return values[0], nil
	}
	result := make(map[string]interface{}, len(values))
	for i, v := range values {
		result[string(m.rows.FieldDescriptions()[i].Name)] = v
	}
	return result, nil
}

func (m *postgresIterator) Close() error {
	m.rows.Close()
	return nil
}

// Parse a file as documents.
// Use an implementation loader based on the filename.
type fileLoader struct {
	path string
}

func (m fileLoader) Iterator(ctx context.Context, _ string) (Iterator, error) {
	f, err := os.Open(m.path)
	if os.IsNotExist(err) {
		return noopLoader{}, nil
	} else if err != nil {
		return nil, err
	}
	if strings.HasSuffix(m.path, ".json.csv") {
		return (&jsonCsvLoader{R: f}).Iterator(ctx, "")
	}
	if strings.HasSuffix(m.path, ".jsonl") {
		return (&jsonLineReader{R: f}).Iterator(ctx, "")
	}
	return (&jsonReaderLoader{R: f}).Iterator(ctx, "")
}

// Read a JSON array into documents.
type jsonReaderLoader struct {
	R io.Reader
}

func (m *jsonReaderLoader) Iterator(_ context.Context, _ string) (Iterator, error) {
	var obj interface{}
	if err := json.NewDecoder(m.R).Decode(&obj); err != nil {
		return nil, err
	}
	if arr, ok := obj.([]interface{}); ok {
		return &memoryIterator{objs: arr}, nil
	} else {
		return &memoryIterator{objs: []interface{}{obj}}, nil
	}
}

// Parse some JSON as a document.
type jsonVerbatimLoader struct{}

func (j jsonVerbatimLoader) Iterator(ctx context.Context, arg string) (Iterator, error) {
	return (&jsonLineReader{R: strings.NewReader(arg)}).Iterator(ctx, "")
}

func NewMemoryIterator(objs []interface{}) Iterator {
	return &memoryIterator{objs: objs}
}

type memoryIterator struct {
	index int
	objs  []interface{}
}

func (m *memoryIterator) Next() bool {
	return m.index < len(m.objs)
}

func (m *memoryIterator) Read(_ context.Context) (interface{}, error) {
	i := m.objs[m.index]
	m.index++
	return i, nil
}

func (m *memoryIterator) Close() error {
	return nil
}

// Read the first cell of each line in the csv into a document.
type jsonCsvLoader struct {
	R io.Reader
}

func (m *jsonCsvLoader) Iterator(_ context.Context, _ string) (Iterator, error) {
	c := csv.NewReader(m.R)
	return &jsonCsvIterator{CR: c}, nil
}

type jsonCsvIterator struct {
	CR  *csv.Reader
	row []string
	err error
}

func (c *jsonCsvIterator) Next() bool {
	c.row, c.err = c.CR.Read()
	return c.err != io.EOF
}

func (c *jsonCsvIterator) Read(_ context.Context) (interface{}, error) {
	if c.err != nil {
		return nil, c.err
	}
	if len(c.row) != 1 {
		return nil, errors.New(".json.csv format must have a single column (the JSON record)")
	}
	var i interface{}
	return i, json.Unmarshal([]byte(c.row[0]), &i)
}

func (c *jsonCsvIterator) Close() error {
	return nil
}

// Read each line of the reader into a document.
type jsonLineReader struct {
	R io.Reader
}

func (m *jsonLineReader) Iterator(_ context.Context, _ string) (Iterator, error) {
	fileScanner := bufio.NewScanner(m.R)
	fileScanner.Split(bufio.ScanLines)
	return &jsonLineIterator{Scanner: fileScanner}, nil
}

type jsonLineIterator struct {
	Scanner *bufio.Scanner
}

func (m *jsonLineIterator) Next() bool {
	return m.Scanner.Scan()
}

func (m *jsonLineIterator) Read(_ context.Context) (interface{}, error) {
	var i interface{}
	return i, json.Unmarshal(m.Scanner.Bytes(), &i)
}

func (m *jsonLineIterator) Close() error {
	return nil
}

type noopLoader struct{}

func (s noopLoader) Iterator(_ context.Context, arg string) (Iterator, error) {
	return s, nil
}

func (s noopLoader) Load(context.Context, string) (interface{}, error) {
	return nil, nil
}

func (s noopLoader) Next() bool {
	return false
}

func (s noopLoader) Read(ctx context.Context) (interface{}, error) {
	return nil, errors.New("nothing to read")
}

func (s noopLoader) Close() error {
	return nil
}
