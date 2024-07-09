package boltpackageInfo

var StructMap_2 map[string]string = map[string]string{
	"DB": `
type DB struct {
	StrictMode bool
	NoSync bool
	NoGrowSync bool
	MmapFlags int
	MaxBatchSize int
	MaxBatchDelay time.Duration
	AllocSize int
	path     string
	file     *os.File
	lockfile *os.File // windows only
	dataref  []byte   // mmap'ed readonly, write throws SEGV
	data     *[maxMapSize]byte
	datasz   int
	filesz   int // current on disk file size
	meta0    *meta
	meta1    *meta
	pageSize int
	opened   bool
	rwtx     *Tx
	txs      []*Tx
	freelist *freelist
	stats    Stats

	pagePool sync.Pool

	batchMu sync.Mutex
	batch   *batch

	rwlock   sync.Mutex   
	metalock sync.Mutex  
	mmaplock sync.RWMutex 
	statlock sync.RWMutex

	ops struct {
		writeAt func(b []byte, off int64) (n int, err error)
	}

	readOnly bool
}`,
	"Tx": `
type Tx struct {
	writable       bool
	managed        bool
	db             *DB
	meta           *meta
	root           Bucket
	pages          map[pgid]*page
	stats          TxStats
	commitHandlers []func()
	WriteFlag int
}`,
	"Info": `type Info struct {
	Data     uintptr
	PageSize int
}`,
	"meta": `type meta struct {
	magic    uint32
	version  uint32
	pageSize uint32
	flags    uint32
	root     bucket
	freelist pgid
	pgid     pgid
	txid     txid
	checksum uint64
}`,
	"page": `type page struct {
	id       pgid
	flags    uint16
	count    uint16
	overflow uint32
	ptr      uintptr
}`,
	"Stats": `
type Stats struct {
	FreePageN     int 
	PendingPageN  int 
	FreeAlloc     int 
	FreelistInuse int 

	TxN     int 
	OpenTxN int 

	TxStats TxStats
},
type TxStats struct {
	PageCount int 
	PageAlloc int 
	CursorCount int 
	NodeCount int 
	NodeDeref int
	Rebalance     int          
	RebalanceTime time.Duration 
	Split     int          
	Spill     int           
	SpillTime time.Duration
	Write     int          
	WriteTime time.Duration 
}`,
	"call": `type call struct {
	fn  func(*Tx) error
	err chan<- error
}`,
	"batch": `type batch struct {
	db    *DB
	timer *time.Timer
	start sync.Once
	calls []call
}`,
	"panicked": `type panicked struct {
	reason interface{}
}`,
	"Options": `
type Options struct {
	Timeout time.Duration
	NoGrowSync bool
	ReadOnly bool
	MmapFlags int
}`}
