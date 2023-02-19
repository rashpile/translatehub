package provider

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Opts struct {
	key          string
	serviceURL   string
	usageURL     string
	query_text   string
	query_target string
	method       string
	authorize    func(req *http.Request) error
	result       func(resType string, resBody []byte) *TranslateResult
	handler      func(reqType string) *TranslateResult
}

func (o *Opts) defaults() {
	if o.query_text == "" {
		o.query_text = "text"
	}
	if o.query_target == "" {
		o.query_target = "target_lang"
	}
	if o.method == "" {
		o.method = "GET"
	}
	if o.usageURL == "" {
		o.usageURL = o.serviceURL + "usage"
	}
}

type Provider interface {
	Translate(text string, sourceLanguage string, targetLanguage string) TranslateResult
	Name() string
	Usage() TranslateResult
}

type TranslateProvider struct {
	name string
	opts Opts
}

type TranslateResult struct {
	Error string
	Text  string
	Usage struct {
		Count int
		Limit int
	}
}

func (h TranslateProvider) Translate(text string, sourceLanguage string, targetLanguage string) TranslateResult {
	h.opts.defaults()
	if h.opts.handler != nil {
		res := h.opts.handler("translate")
		if res != nil {
			return *res
		}
	}
	result := TranslateResult{}
	requestURL := h.opts.serviceURL
	url, err := url.Parse(requestURL)
	if err != nil {
		result.Error = fmt.Errorf("a Provider %s could not parse url: %w", h.name, err).Error()
		return result
	}
	q := url.Query()
	q.Add(h.opts.query_text, text)
	q.Add(h.opts.query_target, targetLanguage)
	url.RawQuery = q.Encode()

	req, err := http.NewRequest(h.opts.method, url.String(), nil)
	if err != nil {
		result.Error = fmt.Errorf("a Provider %s could not create request: %w", h.name, err).Error()
		return result
	}
	if h.opts.authorize != nil {
		if err := h.opts.authorize(req); err != nil {
			result.Error = fmt.Errorf("a Provider %s could not authorize request: %w", h.name, err).Error()
			return result
		}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		result.Error = fmt.Errorf("a Provider %s error making request: %w", h.name, err).Error()
		return result
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		result.Error = fmt.Errorf("a Provider %s error read json: %w", h.name, err).Error()
		return result
	}
	if h.opts.result != nil {
		r := h.opts.result("translate", resBody)
		if r != nil {
			return *r
		}
	}

	return result
}

func (h TranslateProvider) Usage() TranslateResult {
	if h.opts.handler != nil {
		res := h.opts.handler("usage")
		if res != nil {
			return *res
		}
	}

	requestURL := h.opts.usageURL

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return TranslateResult{Error: fmt.Errorf("an Provider %s could not create request: %w", h.name, err).Error()}
	}

	if h.opts.authorize != nil {
		if err := h.opts.authorize(req); err != nil {
			return TranslateResult{Error: fmt.Errorf("a Provider %s could not authorize request: %w", h.name, err).Error()}
		}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return TranslateResult{Error: fmt.Errorf("an Provider %s usage failed: %w", h.name, err).Error()}
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return TranslateResult{Error: fmt.Errorf("an Provider %s failed to read json: %w", h.name, err).Error()}
	}

	// var tr map[string]int
	// if err := json.Unmarshal([]byte(resBody), &tr); err != nil {
	// 	return TranslateResult{error: fmt.Errorf("a Provider %s failed to read response %s: %w", h.name, resBody, err).Error() }
	// }
	if h.opts.result != nil {
		r := h.opts.result("usage", resBody)
		if r != nil {
			return *r
		}
	}
	return TranslateResult{Usage: struct {
		Count int
		Limit int
	}{Count: 0, Limit: 0}}
}

func (h TranslateProvider) Name() string {
	return h.name
}
