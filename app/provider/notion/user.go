package notion

type User struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	Type      string `json:"type,omitempty"`
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Person    *struct {
		Email string `json:"email"`
	} `json:"person,omitempty"`
}
