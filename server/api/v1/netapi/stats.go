package netapi

type Request struct {
	BytesRead      int     `json:"Bytes Read"`
	BytesWritten   int     `json:"Bytes Written"`
	ResponseStatus string  `json:"Response Status"`
	StartTime      float64 `json:"Start Time"`
	EndTime        float64 `json:"End Time"`
	Client         string  `json:"Client"`
	ProcessingTime float64 `json:"Processing Time"`
	RequestLine    string  `json:"Request-Line"`
}

type Applications struct {
	Enabled             bool               `json:"Enabled"`
	BytesReadRequest    float64            `json:"Bytes Read/Request"`
	BytesReadSecond     float64            `json:"Bytes Read/Second"`
	BytesWrittenRequest float64            `json:"Bytes Written/Request"`
	BytesWrittenSecond  float64            `json:"Bytes Written/Second"`
	CurrentTime         float64            `json:"Current Time"`
	CurrentRequests     int                `json:"Current Requests"`
	RequestsSecond      float64            `json:"Requests/Second"`
	ServerVersion       string             `json:"Server Version"`
	StartTime           float64            `json:"Start Time"`
	TotalBytesRead      int                `json:"Total Bytes Read"`
	TotalBytesWritten   int                `json:"Total Bytes Written"`
	TotalRequests       int                `json:"Total Requests"`
	TotalTime           float64            `json:"Total Time"`
	Uptime              float64            `json:"Uptime"`
	Requests            map[string]Request `json:"Requests"`
}

type WorkerThread struct {
	Requests        int     `json:"Requests"`
	BytesRead       int     `json:"Bytes Read"`
	BytesWritten    int     `json:"Bytes Written"`
	WorkTime        float64 `json:"Work Time"`
	ReadThroughput  float64 `json:"Read Throughput"`
	WriteThroughput float64 `json:"Write Throughput"`
}

type HTTPServer struct {
	Enabled         bool                    `json:"Enabled"`
	BindAddress     string                  `json:"Bind Address"`
	RunTime         int                     `json:"Run time"`
	Accepts         int                     `json:"Accepts"`
	AcceptsSec      float64                 `json:"Accepts/sec"`
	Queue           int                     `json:"Queue"`
	Threads         int                     `json:"Threads"`
	ThreadsIdle     int                     `json:"Threads Idle"`
	SocketErrors    int                     `json:"Socket Errors"`
	Requests        int                     `json:"Requests"`
	BytesRead       int                     `json:"Bytes Read"`
	BytesWritten    int                     `json:"Bytes Written"`
	WorkTime        int                     `json:"Work Time"`
	ReadThroughput  float64                 `json:"Read Throughput"`
	WriteThroughput float64                 `json:"Write Throughput"`
	WorkerThreads   map[string]WorkerThread `json:"Worker Threads"`
}

type Stats struct {
	Applications Applications `json:"CherryPy Applications"`
	HTTPServer   HTTPServer   `json:"Cheroot HTTPServer"`
}
