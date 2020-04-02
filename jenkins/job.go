package jenkins

type Job struct {
	Class                 string `json:"_class"`
	Description           string
	DisplayName           string
	DisplayNameOrNull     string
	FullDisplayName       string
	Name                  string
	Url                   string
	Buildable             bool
	Builds                []Build
	Color                 string
	FirstBuild            *Build
	InQueue               bool
	KeepDependencies      bool
	LastBuild             *Build
	LastCompleteBuild     *Build
	LastSuccessfulBuild   *Build
	LastUnstableBuild     *Build
	LastUnsuccessfulBuild *Build
	NextBuildNumber       int32
	ConcurrentBuild       bool
	// TODO: make proper structs for deserializing these
	Actions      []interface{}
	HealthReport []interface{}
	Property     []interface{}
	QueueItem    interface{}
}
