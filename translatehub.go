package translatehub

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rashpile/translatehub/provider"
)

type Request struct {
	SourceLanguage string `json:"sourceLanguage"` // source
	TargetLanguage string `json:"targetLanguage"` // target
	Text           string `json:"text"`
	Engine         string `json:"engine"`
}

type Response struct {
	SourceLanguage string `json:"sourceLanguage"` // source
	TargetLanguage string `json:"targetLanguage"` // target
	Text           string `json:"text"`
	Error          string `json:"error"`
}

type Usage struct {
	Engine  string `json:"engine"`
	Count   int    `json:"count"`
	Limit   int    `json:"limit"`
	Percent string `json:"percent"`
	Message string `json:"message"`
}

type UsageResponse struct {
	Usage []Usage `json:"usage"`
}
type SecretReader interface {
	Get() string
}

type Translate struct {
	providers []provider.Provider
}

func NewTranslate() *Translate {
	return &Translate{}
}

func (t *Translate) Usage() UsageResponse {
	var usage []Usage

	for _, p := range t.providers {
		res := p.Usage()
		ps := 0.0
		if res.Usage.Limit != 0 {
			ps = float64(res.Usage.Count) / float64(res.Usage.Limit) * 100
		}
		usage = append(usage, Usage{
			Engine:  p.Name(),
			Count:   res.Usage.Count,
			Limit:   res.Usage.Limit,
			Percent: fmt.Sprintf("%.2f%%", ps),
			Message: res.Error,
		})
	}
	return UsageResponse{
		Usage: usage,
	}
}

func (t *Translate) Translate(req *Request) Response {
	errors := []string{}
	log.Printf("Engine: %s", req.Engine)
	for _, p := range t.providers {
		if len(req.Engine) == 0 || strings.EqualFold(req.Engine, p.Name()) {
			res := p.Translate(req.Text, req.SourceLanguage, req.TargetLanguage)
			if len(res.Error) == 0 {
				return Response{
					SourceLanguage: req.SourceLanguage,
					TargetLanguage: req.TargetLanguage,
					Text:           res.Text,
				}
			}
			errors = append(errors, fmt.Sprintf("%s: %s", p.Name(), res.Error))
		}
	}
	return Response{
		SourceLanguage: req.SourceLanguage,
		TargetLanguage: req.TargetLanguage,
		Text:           "",
		Error:          strings.Join(errors, "\n"),
	}
}

func (t *Translate) AddProvider(name string, secretReader SecretReader) {
	reader := func() string { return secretReader.Get() }
	switch strings.ToLower(name) {
	case "deepl":
		t.providers = append(t.providers, provider.NewDeepL(reader))
	case "google":
		t.providers = append(t.providers, provider.NewGoogle(reader))
	default:
		return
	}
}
func (t *Translate) ClearProviders() {
	t.providers = []provider.Provider{}
}

func About() {
	fmt.Println("Translate Hub")
}

type FileSecretReader struct {
	Path string
}

func (f *FileSecretReader) Get() string {
	p := f.Path
	if strings.HasPrefix(p, "~") {
		dir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("could not get home dir: %s", err)
		}
		p = dir + p[1:]
	}
	fileRes, err := os.ReadFile(p)
	if err != nil {
		log.Fatalf("could not read file: %s", err)
	}
	return strings.TrimSpace(string(fileRes))
}
