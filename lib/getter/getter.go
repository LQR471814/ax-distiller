package getter

import (
	"crypto/sha1"
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/purell"
	"github.com/huandu/go-sqlbuilder"
	_ "github.com/mattn/go-sqlite3"
)

type Request struct {
	// required
	URL *url.URL
	// if left blank, will default to GET
	Method string
	// can be nil
	Body []byte
	// if not specified, this will be SHA-1 of the body contents
	BodyHash []byte
}

type CachedRequest struct {
	URL      string `db:"url"`
	Method   string `db:"method"`
	BodyHash []byte `db:"body_hash"`
	Response []byte `db:"response"`
}

// A client that does simple HTTP requests with caching
type Getter struct {
	client *http.Client
	db     *sql.DB
}

func NewGetter(cacheLocation string) (Getter, error) {
	client := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   time.Second * 10,
	}

	db, err := sql.Open("sqlite3", cacheLocation)
	if err != nil {
		return Getter{}, err
	}

	builder := sqlbuilder.NewCreateTableBuilder()
	builder.CreateTable("http_cache").IfNotExists()
	builder.Define("url", "TEXT", "NOT NULL")
	builder.Define("method", "TEXT", "NOT NULL")
	builder.Define("body_hash", "BLOB", "NOT NULL")
	builder.Define("response", "BLOB", "NOT NULL")
	builder.Define("PRIMARY", "KEY", "(url, method, body_hash)")
	_, err = db.Exec(builder.String())
	if err != nil {
		return Getter{}, err
	}

	return Getter{
		client: client,
		db:     db,
	}, nil
}

const NORMALIZE_FLAGS = purell.FlagsSafe |
	purell.FlagsUsuallySafeGreedy |
	purell.FlagRemoveFragment

func (g Getter) Do(req Request) ([]byte, error) {
	var err error
	req, err = prepareRequest(req)
	if err != nil {
		return nil, err
	}

	urlString := req.URL.String()

	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("response").
		From("http_cache").
		Where(
			sb.Equal("url", urlString),
			sb.Equal("method", req.Method),
			sb.Equal("body_hash", req.BodyHash),
		).
		Limit(1)
	selectSQL, selectArgs := sb.Build()

	rows, err := g.db.Query(selectSQL, selectArgs...)
	ok := rows.Next()
	if err == nil && ok {
		var cached sql.RawBytes
		err = rows.Scan(&cached)
		if err == nil {
			slog.Info("cache hit", "url", urlString)
			return cached, nil
		} else {
			slog.Warn("failed to query cached request:", "error", err.Error())
		}
	} else if err != nil {
		slog.Warn("failed to query cached request:", "error", err.Error())
	}

	slog.Info("cache miss", "url", urlString)

	res, err := g.client.Do(&http.Request{
		Method: req.Method,
		URL:    req.URL,
	})
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("http_cache")
	ib.Cols("url", "method", "body_hash", "response")
	ib.Values(urlString, req.Method, req.BodyHash, body)
	insertSQL, insertArgs := ib.Build()

	_, err = g.db.Exec(insertSQL, insertArgs...)
	return body, err
}

func prepareRequest(req Request) (Request, error) {
	if req.Method != "" {
		req.Method = strings.ToUpper(req.Method)
	} else {
		req.Method = "GET"
	}

	if req.Body == nil {
		req.BodyHash = make([]byte, 0)
	} else if req.BodyHash == nil {
		hasher := sha1.New()
		hasher.Write(req.Body)
		req.BodyHash = hasher.Sum(nil)
	}

	normalized := purell.NormalizeURL(req.URL, NORMALIZE_FLAGS)
	newURL, err := url.Parse(normalized)
	if err != nil {
		return req, err
	}
	req.URL = newURL
	return req, nil
}
