package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	mimetypes "github.com/whosonfirst/go-whosonfirst-mimetypes"
)

var Links = []string{
	"/trace/",
	"/static/abc.js",
	"/static/abc/xyz.css",
	"/static/abc/xyz/uvw.txt",
	"/static/abc.html",
	"/static/abc.jpg",
	"/dynamic/abc.php",
	"/dynamic/abc.asp",
	"/code/200",
	"/code/400",
	"/code/404",
	"/code/502",
	"/size/11k.zip",
	"/size/1k.bin",
	"/headersize/16k",
	"/slow/3",
	"/slow/4-10",
	"/redirect/301?url=http://www.haiji.pro",
	"/redirect/302?url=http://www.haiji.pro",
	"/redirect/js?url=http://www.haiji.pro",
}

type Server struct {
	tpls             map[string]string
	nodeID           string
	linkHTML         string
	pageTimeFormat   string
	headerTimeFormat string
	runner           *http.Server
}

func (s *Server) Init(addr string, r *mux.Router) {
	s.tpls = make(map[string]string)
	s.tpls["/"] = `
	<h1>YNM30K Test site</h1>
    <h2>Request header</h2>
    <pre>%s
    </pre>
    <h2>Links</h2>
    <ul>%s
    </ul>
    <footer>
        <hr/>SERVER-ID: %s, Powered by YNM30K Fork me on <a href="https://gitee.com/haiji_fsck/ynm30k">Gitee</a>/<a href="https://github.com/HighGee/ynm30k">Github</a>
    </footer>
	`
	s.tpls["/trace/"] = "%s %s %s\r\n%s\r\n\r\n"

	s.nodeID = getNodeId()

	linkLinks := []string{}
	for _, link := range Links {
		linkLinks = append(linkLinks, fmt.Sprintf("<li><a href=\"%s\">%s</a></li>", link, link))
	}
	s.linkHTML = strings.Join(linkLinks, "")

	s.pageTimeFormat = "2006-01-02 15:04:05.999999"
	s.headerTimeFormat = "Mon, 2 Jan 2006 15:04:05 GMT"

	s.runner = &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	headerHtml := getSortedHeaders(r, "<br>")
	fmt.Fprintf(w, s.tpls["/"], headerHtml, s.linkHTML, s.nodeID)
}

func (s *Server) TraceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, s.tpls["/trace/"], r.Method, r.URL, r.Proto, getSortedHeaders(r, "\r\n"))
}

func (s *Server) FileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["file"]
	ext := strings.Split(filename, ".")[1]
	mimetype := strings.Join(mimetypes.TypesByExtension(ext), "\t")
	if mimetype == "" {
		mimetype = strings.Join(mimetypes.TypesByExtension(".html"), "\t")
	}
	w.Header().Set("Content-Type", mimetype)

	cacheTimeInt := 10800
	cacheHeaderStr := r.Header.Get("Cache")
	cacheHeaderInt, err := strconv.Atoi(cacheHeaderStr)
	if err == nil {
		cacheTimeInt = cacheHeaderInt
	}
	now := time.Now().UTC().Add(time.Second * time.Duration(cacheTimeInt))
	w.Header().Set("Expires", now.UTC().Format(s.headerTimeFormat))
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", cacheTimeInt))
	fmt.Fprint(w, r.URL.Path)
}

func (s *Server) DynamicHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	data := make(map[string]interface{})
	header := make(map[string]string)
	header["Host"] = r.Host
	for k, v := range r.Header {
		header[k] = v[0]
	}
	data["body"] = r.Body
	data["uri"] = r.URL.RequestURI()
	data["headers"] = header
	data["path"] = r.URL.Path
	data["query"] = r.URL.RawQuery
	data["path"] = r.URL.Path
	data["arguments"] = r.URL.Query()
	jsonStr, _ := json.MarshalIndent(data, "", "    ")
	UUID := uuid.NewV4()
	fmt.Fprintf(w, "hello :-)<pre>%s<pre><hr>%s", jsonStr, UUID)
}

func (s *Server) CodeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]
	acceptedCodes := map[string]struct{}{
		"200": struct{}{},
		"400": struct{}{},
		"404": struct{}{},
		"502": struct{}{},
	}
	if _, ok := acceptedCodes[code]; !ok {
		code = "404"
	}
	tpl := `<h1>Http %s</h1> <hr/>Generated at %s`
	codeInt, _ := strconv.Atoi(code)
	w.WriteHeader(codeInt)
	fmt.Fprintf(w, tpl, code, time.Now().Format(s.pageTimeFormat))
}

func (s *Server) SizeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sizeName := strings.Split(vars["size"], ".")[0]
	byteCount := 0
	if strings.Contains(sizeName, "k") {
		tmpCount, err := strconv.Atoi(strings.Trim(sizeName, "k"))
		if err == nil {
			byteCount = tmpCount * 1024
		}
	} else if strings.Contains(sizeName, "m") {
		tmpCount, err := strconv.Atoi(strings.Trim(sizeName, "m"))
		if err == nil {
			byteCount = tmpCount * 1024 * 1024
		}
	} else {
		fmt.Fprint(w, "arg err")
		return
	}
	w.Header().Set("Content-Description", "File Transfer")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	ret := make([]byte, byteCount)
	w.Write(ret)
}

func (s *Server) HeaderSizeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sizeName := vars["size"]
	byteCount := 0
	if strings.Contains(sizeName, "k") {
		tmpCount, err := strconv.Atoi(strings.Trim(sizeName, "k"))
		if err == nil {
			byteCount = tmpCount * 1024
		}
	} else if strings.Contains(sizeName, "m") {
		tmpCount, err := strconv.Atoi(strings.Trim(sizeName, "m"))
		if err == nil {
			byteCount = tmpCount * 1024 * 1024
		}
	} else {
		fmt.Fprint(w, "arg err")
		return
	}

	var ret strings.Builder
	for i := 0; i < byteCount; i++ {
		ret.WriteString("f")
	}
	w.Header().Add("Big", ret.String())
	fmt.Fprint(w, "specified header size page")
}

func (s *Server) SlowHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_range := vars["range"]

	start := ""
	end := ""
	pos := strings.Index(_range, "-")
	if pos == -1 {
		start = _range
	} else {
		splits := strings.Split(_range, "-")
		start = splits[0]
		end = splits[1]
	}
	_start, err := strconv.Atoi(start)
	if err != nil {
		fmt.Fprint(w, "err arg")
		return
	}
	_end := 0
	if end != "" {
		_end, _ = strconv.Atoi(end)
		if err != nil {
			fmt.Fprint(w, "err arg")
			return
		}
	}

	i := _start
	if _end != 0 && _end > _start {
		i = rand.Intn(_end-_start) + _start
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprintf(w, "Start at: %s<br/>", time.Now().Format(s.pageTimeFormat))
	time.Sleep(time.Duration(i) * time.Second)
	fmt.Fprintf(w, "End at: %s<br/>", time.Now().Format(s.pageTimeFormat))
}

func (s *Server) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	method := vars["method"]
	args := r.URL.Query()
	url := args["url"][0]
	switch method {
	case "301", "302":
		code, _ := strconv.Atoi(method)
		w.Header().Set("Location", url)
		w.WriteHeader(code)
	case "js":
		fmt.Fprintf(w, "<script>location.href=\"%s\"</script>", url)
	case "meta":
		fmt.Fprintf(w, "<meta http-equiv=\"refresh\" content=\"0; url=%s\" />", url)
	default:
		fmt.Fprint(w, "wrong argument")
	}
}

func (s *Server) Run() {
	s.runner.ListenAndServe()
}
