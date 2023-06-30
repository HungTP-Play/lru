package shared

type MapUrlRequest struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

type MapUrlResponse struct {
	Id        string `json:"id"`
	Url       string `json:"url"`
	Shortened string `json:"shortened"`
}

type RedirectRequest struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

type RedirectResponse struct {
	Id          string `json:"id"`
	Url         string `json:"url"`
	OriginalUrl string `json:"originalUrl"`
}

type MapToRedirectRequest struct {
	Id           string `json:"id"`
	Url          string `json:"url"`
	ShortenedUrl string `json:"shortenedUrl"`
}

type MapToAnalyticsRequest struct {
	Id           string `json:"id"`
	Url          string `json:"url"`
	ShortenedUrl string `json:"shortenedUrl"`
}

type RedirectToAnalyticsRequest struct {
	Id          string `json:"id"`
	Url         string `json:"url"`
	OriginalUrl string `json:"originalUrl"`
}
