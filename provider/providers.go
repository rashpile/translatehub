package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func NewDeepL(secretReader func() string) Provider {
	return TranslateProvider{
		name: "DeepL",
		opts: Opts{
			key:        secretReader(),
			serviceURL: "https://api-free.deepl.com/v2/translate",
			usageURL:   "https://api-free.deepl.com/v2/usage",
			authorize: func(req *http.Request) error {
				req.Header.Add("Authorization", fmt.Sprintf("DeepL-Auth-Key %s", secretReader()))
				return nil
			},
			result: func(resType string, resBody []byte) *TranslateResult {
				switch resType {
				case "usage":
					result := &TranslateResult{}
					var tr map[string]int
					if err := json.Unmarshal(resBody, &tr); err != nil {
						//todo:
						//{"message":"Wrong endpoint. Use https://api-free.deepl.com"}
						result.Error = string(resBody) + " " + err.Error()
						return result
					}
					result.Usage.Count = tr["character_count"]
					result.Usage.Limit = tr["character_limit"]
					return result
				case "translate":
					result := &TranslateResult{}
					var tr map[string][]map[string]string
					if err := json.Unmarshal(resBody, &tr); err != nil {
						result.Error = err.Error()
						return result
					}
					//tr["translations"][0]["detected_source_language"]
					result.Text = tr["translations"][0]["text"]
					return result
				}
				return nil
			},
		},
	}
}

type GoogleTranslateProvider struct {
	TranslateProvider
}

// https://cloud.google.com/translate/docs/basic/translating-text
// https://cloud.google.com/translate/docs/reference/rest/v2/translate
func NewGoogle(secretReader func() string) Provider {
	return TranslateProvider{
		name: "Google",
		opts: Opts{
			key:          secretReader(),
			serviceURL:   "https://translation.googleapis.com/language/translate/v2/",
			query_text:   "q",
			query_target: "target",
			method:       "POST",
			authorize: func(req *http.Request) error {
				q := req.URL.Query()
				q.Add("key", secretReader())
				req.URL.RawQuery = q.Encode()
				return nil
			},
			result: func(resType string, resBody []byte) *TranslateResult {

				switch resType {
				case "usage":
					result := &TranslateResult{}
					var tr map[string]int
					if err := json.Unmarshal(resBody, &tr); err != nil {
						result.Error = err.Error()
						return result
					}
					result.Usage.Count = tr["character_count"]
					result.Usage.Limit = tr["character_limit"]
					return result
				case "translate":
					result := &TranslateResult{}
					var tr map[string]map[string][]map[string]string
					if err := json.Unmarshal([]byte(resBody), &tr); err != nil {
						var tr map[string]string
						if err2 := json.Unmarshal([]byte(resBody), &tr); err2 != nil {
							result.Error = fmt.Errorf("a Provider %s error convert json %s: %w", "google", resBody, err).Error()
						}
						if len(tr["message"]) > 0 {
							result.Error = fmt.Errorf("a Provider %s error: %s", "google", tr["message"]).Error()
						} else {
							result.Error = fmt.Errorf("a Provider %s error convert json '%s': %w", "google", resBody, err).Error()
						}
						return result
					}
					result.Text = tr["data"]["translations"][0]["translatedText"]
					return result
				}
				return nil
			},
			handler: func(reqType string) *TranslateResult {
				res := &TranslateResult{}
				if reqType == "usage" {
					res.Error = "Not implemented: use  link " + "https://console.cloud.google.com/apis/api/translate.googleapis.com/quotas"
					return res
				}
				return nil
			},
		},
	}
}
