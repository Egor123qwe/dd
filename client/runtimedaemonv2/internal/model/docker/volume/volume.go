package volume

type Usage int32

const (
	Local Usage = iota
	Host
	Shared
)

type SharedVolume struct {
	AccessKeyID     string
	SecretAccessKey string

	BucketName string
	Mount      string
}

type SharedVolumeState struct {
	Enabled bool
}
