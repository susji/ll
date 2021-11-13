package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/susji/ll/collection"
)

const DEFAULT_HTML_TEMPLATE = `
<html>
  <body>
    <a rel="noreferrer" href="{{ .url }}">{{ .url }}</a>
  </body>
</html>
`
const (
	reapTime = time.Minute
	dumpTime = time.Minute
)

type server struct {
	logUrls        bool
	renderTemplate string
	decayTime      time.Duration
	decayUses      int
	endpoint       string
	laddr          string
	shortbytes     int
	linkPrefix     string
	dumpFile       string
	schema         map[string]interface{}

	renderer *template.Template
	c        *collection.Collection
}

func urlToMap(u *url.URL) map[string]string {
	return map[string]string{"url": u.String()}
}

func (s *server) fetch(r *http.Request, w http.ResponseWriter, short string) {
	e, last := s.c.Fetch(short)
	if e == nil {
		log.Print("fetch: not found: ", short)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if last {
		log.Print("fetch: decayed due to usage: ", short)
	}

	// We're not even trying to handle this properly. We'll just
	// get the first MIME type which looks like a value we can
	// use.
	var rendererr error
	switch strings.SplitN(r.Header.Get("Accept"), ",", 2)[0] {
	case "text/html":
		w.Header().Add("Content-Type", "text/html")
		rendererr = s.renderer.Execute(w, urlToMap(e.URL))
	case "application/json":
		buf, err := json.Marshal(urlToMap(e.URL))
		if err != nil {
			log.Print("fetch: cannot render as json: ", err)
			rendererr = err
		} else {
			w.Header().Add("Content-Type", "application/json")
			w.Write(buf)
		}
	default:
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte(e.URL.String()))
		rendererr = nil

	}
	if rendererr != nil {
		log.Print("fetch: response rendering failed: ", rendererr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Print("fetch: ", short)
}

func (s *server) submit(w http.ResponseWriter, long string) {
	u, err := url.Parse(long)
	if err != nil {
		log.Print("submit: invalid url: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, ok := s.schema[u.Scheme]; !ok {
		log.Print("submit: unaccepted scheme: ", u.Scheme)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if u.Host == "" {
		log.Print("submit: no host in URL")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortname, err := s.c.Submit(
		u, s.shortbytes, time.Now().Add(s.decayTime), s.decayUses)
	if err != nil {
		log.Print("submit: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if s.logUrls {
		log.Print("submit: ", shortname, " <- ", u)
	} else {
		log.Print("submit: ", shortname)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(
		fmt.Sprintf("%s%s <- %s", s.linkPrefix, shortname, u)))
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) == 0 {
		log.Print("empty request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	raws := strings.SplitN(r.URL.Path[1:], "/", 2)

	if len(raws) == 1 {
		s.fetch(r, w, raws[0])
		return
	}

	if raws[0] != s.endpoint {
		log.Print("unrecognized endpoint: ", raws[0])
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s.submit(w, raws[1])
}

func (s *server) reaper(ctx context.Context, t time.Duration) {
	var cb func(string, *collection.Entry)
	if s.logUrls {
		cb = func(shortname string, e *collection.Entry) {
			log.Print(
				"reaper: decayed ",
				shortname,
				" <- ", e.URL)
		}
	} else {
		cb = func(shortname string, _ *collection.Entry) {
			log.Print("reaper: decayed ", shortname)
		}
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(t):
			s.c.Reap(cb)
		}
	}
}

func (s *server) dump() {
	destfile, err := filepath.Abs(s.dumpFile)
	if err != nil {
		log.Print("dump: ", err)
		return
	}
	tempdir := filepath.Dir(destfile)
	f, err := os.CreateTemp(tempdir, "ll_dump_temp*")
	if err != nil {
		log.Print("dump: ", err)
		return
	}
	s.c.Dump(f)
	f.Close()
	if err := os.Rename(f.Name(), destfile); err != nil {
		log.Print("dump: ", err)
	}
}

func (s *server) dumper(ctx context.Context, t time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(t):
			s.dump()
		}
	}
}

func main() {
	var schema string
	var logTimestamps bool
	s := &server{}
	s.schema = map[string]interface{}{}
	s.c = collection.New()

	flag.BoolVar(
		&logTimestamps,
		"log-timestamps",
		false,
		"Whether to include timestamps in log messages")
	flag.BoolVar(
		&s.logUrls,
		"log-urls",
		false,
		"Whether to log URLs which have been submitted")
	flag.DurationVar(
		&s.decayTime,
		"decay-time",
		time.Hour*24*7,
		"How long from creation until a link decays (0 is infinite)")
	flag.IntVar(
		&s.decayUses,
		"decay-uses",
		0,
		"How many separable accesses decays a link (0 means infinite)")
	flag.StringVar(
		&s.linkPrefix,
		"link-prefix",
		"",
		"Link prefix to display before shortname after submission")
	flag.StringVar(
		&s.endpoint,
		"endpoint",
		"submit",
		"Endpoint for posting new URLs")
	flag.StringVar(
		&schema,
		"accept-schema",
		"https",
		"Comma-separated URI Schema to accept")
	flag.StringVar(&s.laddr, "listen", "localhost:19589", "HTTP listen address")
	flag.IntVar(
		&s.shortbytes,
		"short-length",
		3,
		"Length of shortened URLs")
	flag.StringVar(
		&s.renderTemplate,
		"render-html-template",
		DEFAULT_HTML_TEMPLATE,
		"HTML response template for the shortened URL")
	flag.StringVar(
		&s.dumpFile,
		"dump-file",
		"",
		"File path to dump link data "+
			"(used for initialization if exists during startup)")
	flag.Parse()

	s.renderer = template.Must(
		template.New("renderer").Parse(s.renderTemplate))

	for _, scheme := range strings.Split(schema, ",") {
		s.schema[scheme] = true
	}

	if logTimestamps {
		log.SetFlags(log.LstdFlags)
	} else {
		log.SetFlags(0)
	}

	ctx, cancel := context.WithCancel(context.Background())

	if len(s.dumpFile) > 0 {
		//
		// Three cases here to consider:
		//   - file does not exist or open; it's OK and we move on
		//   - file DOES exist and opens fine; move on
		//   - file DOES exist and does NOT open fine; fatal out
		//
		if _, err := os.Stat(s.dumpFile); err == nil {
			f, err := os.Open(s.dumpFile)
			if err == nil {
				if err := s.c.Import(f); err == nil {
					log.Print(
						"Imported collection from ",
						s.dumpFile)
				} else {
					log.Fatal(
						"Failed importing collection: ", err)
				}
			}
		}
		go s.dumper(ctx, dumpTime)
		}

	srv := http.Server{Addr: s.laddr, Handler: s}
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	go func() {
		<-sigint
		log.Print("SIGINT received")
		cancel()
		srv.Shutdown(ctx)
	}()
	go s.reaper(ctx, reapTime)

	log.Print("Submission endpoint... ", s.endpoint)
	log.Print("URL logging........... ", s.logUrls)
	log.Print("Listen address........ ", s.laddr)
	log.Print("Decay time............ ", s.decayTime)
	log.Print("Decay uses............ ", s.decayUses)
	log.Print("Link prefix........... ", s.linkPrefix)
	log.Print("Data dump file........ ", s.dumpFile)
	if err := srv.ListenAndServe(); err != nil {
		log.Print(err)
	}
	if len(s.dumpFile) > 0 {
		s.dump()
	}
}
