package helpers

type BatchSaver interface {
	Save([]Incoming, string) ([]Output, error)
}

type Output struct {
	OriginalURL string `json:"correlation_id"`
	ShortURL    string `json:"short_url"`
}

type Incoming struct {
	ShortURL    string `json:"correlation_id"`
	OriginalURL string `json:"original_url"`
}
