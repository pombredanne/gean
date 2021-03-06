package data

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/govenue/assert"
	"github.com/govenue/configurator"
	"github.com/govenue/require"
)

func TestGetCSV(t *testing.T) {
	t.Parallel()

	ns := New(newDeps(configurator.New()))

	for i, test := range []struct {
		sep     string
		url     string
		content string
		expect  interface{}
	}{
		// Remotes
		{
			",",
			`http://success/`,
			"gomeetup,city\nyes,Sydney\nyes,San Francisco\nyes,Stockholm\n",
			[][]string{{"gomeetup", "city"}, {"yes", "Sydney"}, {"yes", "San Francisco"}, {"yes", "Stockholm"}},
		},
		{
			",",
			`http://error.extra.field/`,
			"gomeetup,city\nyes,Sydney\nyes,San Francisco\nyes,Stockholm,EXTRA\n",
			false,
		},
		{
			",",
			`http://error.no.sep/`,
			"gomeetup;city\nyes;Sydney\nyes;San Francisco\nyes;Stockholm\n",
			false,
		},
		{
			",",
			`http://nofound/404`,
			``,
			false,
		},

		// Locals
		{
			";",
			"pass/semi",
			"gomeetup;city\nyes;Sydney\nyes;San Francisco\nyes;Stockholm\n",
			[][]string{{"gomeetup", "city"}, {"yes", "Sydney"}, {"yes", "San Francisco"}, {"yes", "Stockholm"}},
		},
		{
			";",
			"fail/no-file",
			"",
			false,
		},
	} {
		msg := fmt.Sprintf("Test %d", i)

		// Setup HTTP test server
		var srv *httptest.Server
		srv, ns.client = getTestServer(func(w http.ResponseWriter, r *http.Request) {
			if !haveHeader(r.Header, "Accept", "text/csv") && !haveHeader(r.Header, "Accept", "text/plain") {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			if r.URL.Path == "/404" {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			w.Header().Add("Content-type", "text/csv")

			w.Write([]byte(test.content))
		})
		defer func() { srv.Close() }()

		// Setup local test file for schema-less URLs
		if !strings.Contains(test.url, ":") && !strings.HasPrefix(test.url, "fail/") {
			f, err := ns.deps.Fs.Source.Create(filepath.Join(ns.deps.Cfg.GetString("workingDir"), test.url))
			require.NoError(t, err, msg)
			f.WriteString(test.content)
			f.Close()
		}

		// Get on with it
		got, err := ns.GetCSV(test.sep, test.url)

		if _, ok := test.expect.(bool); ok {
			assert.Error(t, err, msg)
			continue
		}
		require.NoError(t, err, msg)
		require.NotNil(t, got, msg)

		assert.EqualValues(t, test.expect, got, msg)
	}
}

func TestGetJSON(t *testing.T) {
	t.Parallel()

	ns := New(newDeps(configurator.New()))

	for i, test := range []struct {
		url     string
		content string
		expect  interface{}
	}{
		{
			`http://success/`,
			`{"gomeetup":["Sydney","San Francisco","Stockholm"]}`,
			map[string]interface{}{"gomeetup": []interface{}{"Sydney", "San Francisco", "Stockholm"}},
		},
		{
			`http://malformed/`,
			`{gomeetup:["Sydney","San Francisco","Stockholm"]}`,
			false,
		},
		{
			`http://nofound/404`,
			``,
			false,
		},
		// Locals
		{
			"pass/semi",
			`{"gomeetup":["Sydney","San Francisco","Stockholm"]}`,
			map[string]interface{}{"gomeetup": []interface{}{"Sydney", "San Francisco", "Stockholm"}},
		},
		{
			"fail/no-file",
			"",
			false,
		},
	} {
		msg := fmt.Sprintf("Test %d", i)

		// Setup HTTP test server
		var srv *httptest.Server
		srv, ns.client = getTestServer(func(w http.ResponseWriter, r *http.Request) {
			if !haveHeader(r.Header, "Accept", "application/json") {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			if r.URL.Path == "/404" {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			w.Header().Add("Content-type", "application/json")

			w.Write([]byte(test.content))
		})
		defer func() { srv.Close() }()

		// Setup local test file for schema-less URLs
		if !strings.Contains(test.url, ":") && !strings.HasPrefix(test.url, "fail/") {
			f, err := ns.deps.Fs.Source.Create(filepath.Join(ns.deps.Cfg.GetString("workingDir"), test.url))
			require.NoError(t, err, msg)
			f.WriteString(test.content)
			f.Close()
		}

		// Get on with it
		got, err := ns.GetJSON(test.url)

		if _, ok := test.expect.(bool); ok {
			assert.Error(t, err, msg)
			continue
		}
		require.NoError(t, err, msg)
		require.NotNil(t, got, msg)

		assert.EqualValues(t, test.expect, got, msg)
	}
}

func TestParseCSV(t *testing.T) {
	t.Parallel()

	for i, test := range []struct {
		csv []byte
		sep string
		exp string
		err bool
	}{
		{[]byte("a,b,c\nd,e,f\n"), "", "", true},
		{[]byte("a,b,c\nd,e,f\n"), "~/", "", true},
		{[]byte("a,b,c\nd,e,f"), "|", "a,b,cd,e,f", false},
		{[]byte("q,w,e\nd,e,f"), ",", "qwedef", false},
		{[]byte("a|b|c\nd|e|f|g"), "|", "abcdefg", true},
		{[]byte("z|y|c\nd|e|f"), "|", "zycdef", false},
	} {
		msg := fmt.Sprintf("Test %d: %v", i, test)

		csv, err := parseCSV(test.csv, test.sep)
		if test.err {
			assert.Error(t, err, msg)
			continue
		}
		require.NoError(t, err, msg)

		act := ""
		for _, v := range csv {
			act = act + strings.Join(v, "")
		}

		assert.Equal(t, test.exp, act, msg)
	}
}

func haveHeader(m http.Header, key, needle string) bool {
	var s []string
	var ok bool

	if s, ok = m[key]; !ok {
		return false
	}

	for _, v := range s {
		if v == needle {
			return true
		}
	}
	return false
}
