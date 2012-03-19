package goairbrake

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendNotice(t *testing.T) {
	ApiKey = "a247154ec9993ed53448243555e9479f"
	ts := httptest.NewServer(Watch(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("ahh... no!!!")
	})))
	defer ts.Close()

	SentNotice = make(chan *Noticed)
	http.Get(ts.URL)
	nd := <-SentNotice
	if nd.Url == "" {
		t.Errorf("didn't create airbrake notice, %+v", nd)
	}

}

func TestUnmarshalXml(t *testing.T) {
	x := `
<notice>
  <error-id type="integer">40896325</error-id>
  <url>http://theplant.airbrake.io/errors/40896325/notices/4791302129</url>
  <id type="integer">4791302129</id>
</notice>
`
	var nd *Noticed
	xml.Unmarshal([]byte(x), &nd)
	if nd.ErrorId != 40896325 {
		t.Errorf("wrong unmarshall", nd)
	}
}
