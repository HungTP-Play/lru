package dto

type ShortenRequestDto struct {
	Url string `json:"url"`
}

type ShortenResponseDto struct {
	Url        string `json:"url"`
	ShortedUrl string `json:"shortedUrl"`
}

type RedirectResponseDto struct {
	Url         string `json:"url"`
	OriginalUrl string `json:"originalUrl"`
}
