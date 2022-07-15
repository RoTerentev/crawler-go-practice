package internal

type Config struct {
	WorkersAmount  int `json:"workersAmnt"`
	PagesPerSecond int `json:"pagePerSecond"`
}
