package jenkins

type Build struct {
	Class  string `json:"_class"`
	Number int32  `json:"number"`
	Url    string `json:"url"`
}
