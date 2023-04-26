package restytool

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"

	"sync"
)

// LonelyClient 孤独的客户端，不设置baseurl
type LonelyClient struct {
	httpClient *resty.Client
}

func (itself *LonelyClient) SetProxy(proxyURL string) *LonelyClient {
	itself.httpClient.SetProxy(proxyURL)
	return itself
}

var LonelyOnce sync.Once

var LonelyStd = LonelyClient{}

func StdLonelyClient() *LonelyClient {
	LonelyOnce.Do(func() {
		client := resty.New()
		LonelyStd.httpClient = client
	})
	return &LonelyStd
}

func (itself *LonelyClient) Get(uri string, data ...map[string]string) (resp *resty.Response, err error) {
	queryData := map[string]string{}
	if len(data) == 1 {
		queryData = data[0]
	}
	return itself.httpClient.R().SetQueryParams(queryData).Get(uri)
}

func (itself *LonelyClient) Post(uri string, data ...any) (resp *resty.Response, err error) {
	var bodyData any
	if len(data) == 1 {
		bodyData = data[0]
	}
	return itself.httpClient.R().SetBody(bodyData).Post(uri)
}
func (itself *LonelyClient) PostFormData(uri string, data ...map[string]string) (resp *resty.Response, err error) {
	formData := map[string]string{}
	if len(data) == 1 {
		formData = data[0]
	}
	return itself.httpClient.R().SetFormData(formData).Post(uri)
}

func (itself *LonelyClient) downHttpFile(url string, filename string) (*resty.Response, error) {
	return itself.httpClient.R().SetOutput(filename).Get(url)
}

func DownFile(url string, filename string) (*resty.Response, error) {
	return StdLonelyClient().downHttpFile(url, filename)
}

func GetCurlByR(r resty.Response) bytes.Buffer {
	b2 := bytes.Buffer{}
	b2.WriteString(fmt.Sprintf("curl '%v' -X '%v'", r.Request.URL, r.Request.Method))
	for header, headerValue := range r.Request.Header {
		b2.WriteString(fmt.Sprintf(" -H '%v:%v'", header, headerValue[len(headerValue)-1]))
	}
	if r.Request.Body != nil {
		body, _ := json.Marshal(r.Request.Body)
		b2.WriteString(fmt.Sprintf(" --data-raw '%v' ", string(body)))
	}
	if r.Request.FormData != nil {
		for key, value := range r.Request.FormData {
			b2.WriteString(fmt.Sprintf(` --form '%v="%v" `, key, value))
		}
	}
	b2.WriteString(" --compressed --insecure ")
	return b2
}
