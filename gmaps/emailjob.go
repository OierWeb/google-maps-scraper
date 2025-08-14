package gmaps

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/gosom/google-maps-scraper/exiter"
	"github.com/gosom/scrapemate"
	"github.com/mcnijman/go-emailaddress"
	"github.com/playwright-community/playwright-go"
)

type EmailExtractJobOptions func(*EmailExtractJob)

type EmailExtractJob struct {
	scrapemate.Job

	Entry       *Entry
	ExitMonitor exiter.Exiter
}

func NewEmailJob(parentID string, entry *Entry, opts ...EmailExtractJobOptions) *EmailExtractJob {
	const (
		defaultPrio       = scrapemate.PriorityHigh
		defaultMaxRetries = 0
	)

	job := EmailExtractJob{
		Job: scrapemate.Job{
			ID:         uuid.New().String(),
			ParentID:   parentID,
			Method:     "GET",
			URL:        entry.WebSite,
			MaxRetries: defaultMaxRetries,
			Priority:   defaultPrio,
		},
	}

	job.Entry = entry

	for _, opt := range opts {
		opt(&job)
	}

	return &job
}

func WithEmailJobExitMonitor(exitMonitor exiter.Exiter) EmailExtractJobOptions {
	return func(j *EmailExtractJob) {
		j.ExitMonitor = exitMonitor
	}
}

func (j *EmailExtractJob) Process(ctx context.Context, resp *scrapemate.Response) (any, []scrapemate.IJob, error) {
	defer func() {
		resp.Document = nil
		resp.Body = nil
	}()

	defer func() {
		if j.ExitMonitor != nil {
			j.ExitMonitor.IncrPlacesCompleted(1)
		}
	}()

	log := scrapemate.GetLoggerFromContext(ctx)

	log.Info("Processing email job", "url", j.URL)

	// if html fetch failed just return
	if resp.Error != nil {
		return j.Entry, nil, nil
	}

	doc, ok := resp.Document.(*goquery.Document)
	if !ok {
		return j.Entry, nil, nil
	}

	emails := docEmailExtractor(doc)
	if len(emails) == 0 {
		emails = regexEmailExtractor(resp.Body)
	}

	j.Entry.Emails = emails

	return j.Entry, nil, nil
}

func (j *EmailExtractJob) ProcessOnFetchError() bool {
	return true
}

// GetURL devuelve la URL completa para el trabajo de correo electrónico
func (j *EmailExtractJob) GetURL() string {
	return j.URL
}

// BrowserActions implementa la interfaz scrapemate.IJob para EmailExtractJob
// con un timeout más largo para sitios web lentos o no respondientes
func (j *EmailExtractJob) BrowserActions(ctx context.Context, page playwright.Page) scrapemate.Response {
	var resp scrapemate.Response
	// Aumentamos el timeout a 3 minutos (180000ms) para sitios web lentos
	const timeout = 180000

	// Configuramos un contexto con timeout para toda la operación
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeout+5000)*time.Millisecond)
	defer cancel()

	// Canal para manejar la finalización de la operación
	done := make(chan struct{})
	var pageResponse playwright.Response
	var err error

	// Ejecutamos la navegación en una goroutine
	go func() {
		defer close(done)
		// Intentamos navegar a la página con un timeout extendido
		pageResponse, err = page.Goto(j.GetURL(), playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateLoad,
			Timeout:   playwright.Float(timeout),
		})
	}()

	// Esperamos a que termine la navegación o se agote el tiempo
	select {
	case <-ctxWithTimeout.Done():
		// Si el contexto se cancela, registramos un error de timeout
		resp.Error = fmt.Errorf("timeout excedido al cargar %s", j.URL)
		return resp
	case <-done:
		// La navegación ha terminado (con éxito o error)
		if err != nil {
			// Si hay un error, intentamos capturar el contenido de la página de todos modos
			resp.Error = err
			
			// Intentamos obtener el contenido actual de la página, incluso si hubo error
			body, contentErr := page.Content()
			if contentErr == nil && body != "" {
				// Si pudimos obtener contenido, lo usamos a pesar del error de navegación
				resp.Body = []byte(body)
				resp.URL = page.URL()
				resp.StatusCode = http.StatusOK // Asumimos OK ya que tenemos contenido
			}
			return resp
		}
	}

	// Si llegamos aquí, la navegación fue exitosa
	resp.URL = pageResponse.URL()
	resp.StatusCode = pageResponse.Status()
	resp.Headers = make(http.Header, len(pageResponse.Headers()))

	for k, v := range pageResponse.Headers() {
		resp.Headers.Add(k, v)
	}

	// Obtenemos el contenido de la página
	body, err := page.Content()
	if err != nil {
		resp.Error = err
		return resp
	}

	resp.Body = []byte(body)
	return resp
}

func docEmailExtractor(doc *goquery.Document) []string {
	seen := map[string]bool{}

	var emails []string

	doc.Find("a[href^='mailto:']").Each(func(_ int, s *goquery.Selection) {
		mailto, exists := s.Attr("href")
		if exists {
			value := strings.TrimPrefix(mailto, "mailto:")
			if email, err := getValidEmail(value); err == nil {
				if !seen[email] {
					emails = append(emails, email)
					seen[email] = true
				}
			}
		}
	})

	return emails
}

func regexEmailExtractor(body []byte) []string {
	seen := map[string]bool{}

	var emails []string

	addresses := emailaddress.Find(body, false)
	for i := range addresses {
		if !seen[addresses[i].String()] {
			emails = append(emails, addresses[i].String())
			seen[addresses[i].String()] = true
		}
	}

	return emails
}

func getValidEmail(s string) (string, error) {
	email, err := emailaddress.Parse(strings.TrimSpace(s))
	if err != nil {
		return "", err
	}

	return email.String(), nil
}
