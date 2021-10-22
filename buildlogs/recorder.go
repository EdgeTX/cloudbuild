package buildlogs

type Recorder struct {
	stdOut string
	stdErr string
}

func NewRecorder() *Recorder {
	return &Recorder{
		stdOut: "",
		stdErr: "",
	}
}

func (recorder *Recorder) AddStdOut(data string) {
	recorder.stdOut += data
}

func (recorder *Recorder) AddStdErr(data string) {
	recorder.stdErr += data
}

func (recorder *Recorder) Logs() string {
	return recorder.stdOut + recorder.stdErr
}
