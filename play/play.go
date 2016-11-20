//simpol-online
package play

import (
	"appengine"
	"appengine/datastore"
	"crypto/sha1"
	"fmt"
	"github.com/edwindvinas/simpol/parser"
	"github.com/edwindvinas/simpol/vm"
	"html/template"
	"net/http"

	simpol_core "github.com/edwindvinas/simpol/builtins"
	simpol_encoding_json "github.com/edwindvinas/simpol/builtins/encoding/json"
	simpol_flag "github.com/edwindvinas/simpol/builtins/flag"
	simpol_fmt "github.com/edwindvinas/simpol/builtins/fmt"
	simpol_io "github.com/edwindvinas/simpol/builtins/io"
	simpol_io_ioutil "github.com/edwindvinas/simpol/builtins/io/ioutil"
	simpol_math "github.com/edwindvinas/simpol/builtins/math"
	simpol_math_rand "github.com/edwindvinas/simpol/builtins/math/rand"
	simpol_net "github.com/edwindvinas/simpol/builtins/net"
	simpol_net_http "github.com/edwindvinas/simpol/builtins/net/http"
	simpol_net_url "github.com/edwindvinas/simpol/builtins/net/url"
	simpol_os "github.com/edwindvinas/simpol/builtins/os"
	simpol_os_exec "github.com/edwindvinas/simpol/builtins/os/exec"
	simpol_path "github.com/edwindvinas/simpol/builtins/path"
	simpol_path_filepath "github.com/edwindvinas/simpol/builtins/path/filepath"
	simpol_regexp "github.com/edwindvinas/simpol/builtins/regexp"
	simpol_sort "github.com/edwindvinas/simpol/builtins/sort"
	simpol_strings "github.com/edwindvinas/simpol/builtins/strings"
	simpol_time "github.com/edwindvinas/simpol/builtins/time"

	simpol_colortext "github.com/edwindvinas/simpol/builtins/github.com/daviddengcn/go-colortext"
)

type Record struct {
	Code string
}

var t = template.Must(template.ParseFiles("tmpl/index.tpl"))

func init() {
	http.HandleFunc("/api/play", serveApiPlay)
	http.HandleFunc("/api/save", serveApiSave)
	http.HandleFunc("/p/", servePermalink)
	http.HandleFunc("/", servePermalink)
}

func serveApiSave(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("serveApiSave()")
	code := r.FormValue("code")
	h := sha1.New()
	fmt.Fprintf(h, "%s", code)
	hid := fmt.Sprintf("%x", h.Sum(nil))
	key := datastore.NewKey(c, "Simpol", hid, 0, nil)
	_, err := datastore.Put(c, key, &Record{code})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", key.StringID())
}

func serveApiPlay(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("serveApiPlay()")
	code := r.FormValue("code")
	isDebug := r.FormValue("debug")
	c.Infof("debug: %v", isDebug)
	scanner := new(parser.Scanner)
	c.Infof("serveApiPlay():parser.Scanner")
	scanner.Init(code)
	stmts, err := parser.Parse(w,r,scanner)
	c.Infof("serveApiPlay():parser.Parse:stmts: %v", stmts)
	if e, ok := err.(*parser.Error); ok {
		w.WriteHeader(500)
		fmt.Fprintf(w, "%d: %s\n", e.Pos.Line, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	env := vm.NewEnv()

	c.Infof("serveApiPlay():simpol_core.Import")
	simpol_core.Import(env)

	tbl := map[string]func(env *vm.Env) *vm.Env{
		"encoding/json": simpol_encoding_json.Import,
		"flag":          simpol_flag.Import,
		"fmt":           simpol_fmt.Import,
		"io":            simpol_io.Import,
		"io/ioutil":     simpol_io_ioutil.Import,
		"math":          simpol_math.Import,
		"math/rand":     simpol_math_rand.Import,
		"net":           simpol_net.Import,
		"net/http":      simpol_net_http.Import,
		"net/url":       simpol_net_url.Import,
		"os":            simpol_os.Import,
		"os/exec":       simpol_os_exec.Import,
		"path":          simpol_path.Import,
		"path/filepath": simpol_path_filepath.Import,
		"regexp":        simpol_regexp.Import,
		"sort":          simpol_sort.Import,
		"strings":       simpol_strings.Import,
		"time":          simpol_time.Import,
		"github.com/daviddengcn/go-colortext": simpol_colortext.Import,
	}

	env.Define("import", func(s string) interface{} {
		if loader, ok := tbl[s]; ok {
			return loader(env)
		}
		panic(fmt.Sprintf("package '%s' not found", s))
	})

	env.Define("println", func(a ...interface{}) {
		fmt.Fprint(w, fmt.Sprintln(a...))
	})
	env.Define("print", func(a ...interface{}) {
		fmt.Fprint(w, fmt.Sprint(a...))
	})
	env.Define("prinf", func(a string, b ...interface{}) {
		fmt.Fprintf(w, fmt.Sprintf(a, b...))
	})
	env.Define("panic", func(a ...interface{}) {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Can't use panic()")
		return
	})
	env.Define("load", func(a ...interface{}) {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Can't use load()")
		return
	})
	defer env.Destroy()
	c.Infof("serveApiPlay():vm.Run")
	_, err = vm.Run(stmts, env)
	if err != nil {
		w.WriteHeader(500)
		if e, ok := err.(*vm.Error); ok {
			fmt.Fprintf(w, "%d: %s\n", e.Pos.Line, err)
		} else if e, ok := err.(*parser.Error); ok {
			fmt.Fprintf(w, "%d: %s\n", e.Pos.Line, err)
		} else {
			fmt.Fprintln(w, e.Error())
		}
	}
}

func servePermalink(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof("servePermalink()")
	path := r.URL.Path
	var code string
	if len(path) > 3 {
		id := path[3:]
		c := appengine.NewContext(r)
		var record Record
		err := datastore.Get(c, datastore.NewKey(c, "Simpol", id, 0, nil), &record)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		code = record.Code
	} else {
		code = `var fmt = import("fmt")

println(fmt.Sprintf("こんにちわ世界 %05d", 123))`
	}

	err := t.Execute(w, &struct{ Code string }{Code: code})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}