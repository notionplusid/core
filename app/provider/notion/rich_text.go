package notion

// RichText representation.
type RichText struct {
	PlainText   string `json:"plain_text"`
	HRef        string `json:"href,omitempty"`
	Annotations struct {
		Bold          bool   `json:"bool"`
		Italic        bool   `json:"italic"`
		Strikethrough bool   `json:"strikethrough"`
		Underline     bool   `json:"underline"`
		Code          bool   `json:"code"`
		Color         string `json:"color"`
	} `json:"annotations"`
	Type string `json:"type"`
	Text *struct {
		Content string `json:"content"`
		Link    *struct {
			URL string `json:"url,omitempty"`
		} `json:"link,omitempty"`
	} `json:"text,omitempty"`
	Mention *struct {
		Type string `json:"type"`
		User *User  `json:"user,omitempty"`
		Page *struct {
			ID string `json:"id"`
		} `json:"page,omitempty"`
		Database *struct {
			ID string `json:"id"`
		} `json:"database,omitempty"`
		Date *PagePropertyDate `json:"date,omitempty"`
	} `json:"mention,omitempty"`
	Equation *struct {
		Expression string `json:"expression"`
	} `json:"equation,omitempty"`
}
