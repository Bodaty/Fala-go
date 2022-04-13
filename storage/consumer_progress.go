package storage

// ConsumerProgress reads and writes the last processed index of the job in the job queue
type ConsumerProgress interface {
	// read the current processed index
	ProcessedIndex() (uint64, error)
	// insert the default processed index to the storage layer, can only be done once.
	// initialize for the second time will return storage.ErrAlreadyExists
	InitProcessedIndex(defaultIndex uint64) error
	// update the processed index in the storage layer.
	// it will fail if InitProcessedIndex was never called.
	SetProcessedIndex(processed uint64) error

	// read the current halted status
	Halted() (error, error)
	// insert nil value entry for halted into the storage layer, can only be done once.
	// initialize for the second time will return storage.ErrAlreadyExists
	InitHalted() error
	// update the halted status in the storage layer.
	// it will fail if InitHalted was never called.
	SetHalted(err error) error
}
