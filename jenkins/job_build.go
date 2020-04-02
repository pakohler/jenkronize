package jenkins

type JobBuild struct {
	Class             string `json:"_class"`
	Actions           []interface{}
	Artifacts         []*Artifact
	Building          bool
	Description       string
	DisplayName       string
	Duration          int64
	EstimatedDuration int64
	Executor          interface{}
	FullDisplayName   string
	Id                string
	KeepLog           bool
	Number            int32
	QueueId           int32
	Result            string
	Timestamp         int64
	Url               string
	ChangeSets        []interface{}
	NextBuild         *Build
	PreviousBuild     *Build
}
