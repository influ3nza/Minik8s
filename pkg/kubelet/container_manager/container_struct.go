package container_manager

type VolumeMap struct {
	Host_      string
	Container_ string
	Subdir_    string
	Type_      string
}

var IDLength int = 64
