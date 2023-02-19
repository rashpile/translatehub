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
}

type Response struct {
	SourceLanguage string `json:"sourceLanguage"` // source
	TargetLanguage string `json:"targetLanguage"` // target
	Text           string `json:"text"`
	Error          string `json:"error"`
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

func (t *Translate) Usage() string {
	var s []string
	for _, p := range t.providers {
		res := p.Usage()
		if len(res.Error) > 0 {
			s = append(s, fmt.Sprintf("%s: %s", p.Name(), res.Error))
		} else {
			s = append(s, fmt.Sprintf("%s - %.2f%%", p.Name(), float64(res.Usage.Count)/float64(res.Usage.Limit)*100))
		}
	}
	return strings.Join(s, "\n")
}

func (t *Translate) Translate(req *Request) Response {
	errors := []string{}
	for _, p := range t.providers {
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
