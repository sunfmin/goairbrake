package goairbrake

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
)

var ApiKey string
var ApiNoticeURL = "http://airbrake.io/notifier_api/v2/notices/"
var Environment string

type Notifier struct {
	Name    string `xml:"name"`
	Version string `xml:"version"`
	Url     string `xml:"url"`
}

type BacktraceLine struct {
	Number int    `xml:"number,attr"`
	File   string `xml:"file,attr"`
	Method string `xml:"method,attr"`
}

type Error struct {
	Class     string           `xml:"class"`
	Message   string           `xml:"message"`
	Backtrace []*BacktraceLine `xml:"backtrace>line"`
}

type KeyValue struct {
	Key   string `xml:"key,attr"`
	Value string `xml:",chardata"`
}

type Request struct {
	Url       string      `xml:"url"`
	Component string      `xml:"component"`
	Action    string      `xml:"action"`
	CgiData   []*KeyValue `xml:"cgi-data>var"`
}

type ServerEnvironment struct {
	ProjectRoot     string `xml:"project-root"`
	EnvironmentName string `xml:"environment-name"`
	Hostname        string `xml:"hostname"`
}

type Notice struct {
	XMLName xml.Name `xml:"notice"`

	ApiKey            string             `xml:"api-key"`
	Notifier          *Notifier          `xml:"notifier"`
	Error             *Error             `xml:"error"`
	Request           *Request           `xml:"request"`
	ServerEnvironment *ServerEnvironment `xml:"server-environment"`
	Version           string             `xml:"version,attr"`
}

func (r *Request) AddCgiKeyValue(key string, value string) {
	if r.CgiData == nil {
		r.CgiData = []*KeyValue{}
	}
	r.CgiData = append(r.CgiData, &KeyValue{key, value})
	return
}

func (e *Error) AddBacktrace(number int, file string, method string) {
	if e.Backtrace == nil {
		e.Backtrace = []*BacktraceLine{}
	}
	e.Backtrace = append(e.Backtrace, &BacktraceLine{Number: number, File: file, Method: method})
	return
}

func NewNotice() (n *Notice) {
	n = &Notice{}
	n.Version = "2.2"
	n.ApiKey = ApiKey
	n.Notifier = &Notifier{
		Name:    "goairbrake",
		Version: "0.1.0",
		Url:     "http://github.com/sunfmin/goairbrake",
	}
	pr, _ := os.Getwd()
	hn, _ := os.Hostname()
	n.ServerEnvironment = &ServerEnvironment{
		Hostname:        hn,
		ProjectRoot:     pr,
		EnvironmentName: Environment,
	}
	return n
}

func (n *Notice) SetError(err interface{}) {
	if n.Error == nil {
		n.Error = &Error{Message: fmt.Sprintf("%v", err)}
	}
	for i := 2; ; i++ { // Caller we care about is the user, 2 frames up
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		f := runtime.FuncForPC(pc)
		n.Error.AddBacktrace(line, file, f.Name())
	}
}

func (n *Notice) SetValueFromRequest(req *http.Request) {
	if n.Request == nil {
		n.Request = &Request{}
	}
	n.Request.Url = req.URL.String()
	n.Request.AddCgiKeyValue("SERVER_SOFTWARE", "go")
	n.Request.AddCgiKeyValue("PATH_INFO", req.URL.Path)
	n.Request.AddCgiKeyValue("HTTP_HOST", req.Host)
	n.Request.AddCgiKeyValue("GATEWAY_INTERFACE", "CGI/1.1")
	n.Request.AddCgiKeyValue("REQUEST_METHOD", req.Method)
	n.Request.AddCgiKeyValue("QUERY_STRING", req.URL.RawQuery)
	n.Request.AddCgiKeyValue("REQUEST_URI", req.URL.RequestURI())
	n.Request.AddCgiKeyValue("REMOTE_ADDR", req.RemoteAddr)
	n.Request.AddCgiKeyValue("REMOTE_HOST", req.RemoteAddr)
	n.Request.AddCgiKeyValue("HTTP_ACCEPT", req.Header.Get("Accept-Encoding"))
	n.Request.AddCgiKeyValue("HTTP_USER_AGENT", req.UserAgent())
	env := os.Environ()
	for _, val := range env {
		pair := strings.Split(val, "=")
		if len(pair) < 2 {
			continue
		}
		if strings.Index(pair[0], "_KEY") >= 0 {
			continue
		}
		n.Request.AddCgiKeyValue(pair[0], pair[1])
	}
}
