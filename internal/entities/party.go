package entities

type Party struct {
	PartySize int    `json:"party_size"`
	Name      string `json:"name"`
	Token     string `json:"token"`
}
