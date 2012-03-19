package goairbrake

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
)

type Noticed struct {
	XMLName xml.Name `xml:"notice"`

	ErrorId int64  `xml:"error-id"`
	Url     string `xml:"url"`
	Id      int64  `xml:"id"`
}

var SentNotice chan *Noticed

func Watch(hf http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			notice := NewNotice()
			notice.SetError(err)
			notice.SetValueFromRequest(r)

			go func() {
				generated, _ := xml.Marshal(notice)
				bytexml := bytes.NewBufferString(xml.Header)
				bytexml.Write(generated)

				res, err := http.Post(ApiNoticeURL, "text/xml", bytexml)
				if err != nil {
					log.Println("Post to airbrake.io error: ", err)
				}
				b, _ := ioutil.ReadAll(res.Body)
				if res.StatusCode != 200 {
					log.Println("Post to airbrake.io status error: ", string(b))
				}

				if SentNotice != nil {
					var nd *Noticed
					xml.Unmarshal(b, &nd)
					SentNotice <- nd
				}
			}()
		}()
		hf(w, r)

	}
}
