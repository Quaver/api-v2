package tasks

// Task The task that will be run by the workers.
// The task takes any number of inputs and spits out an output.
type Task func(input ...interface{}) (interface{}, error)

// Worker The individual goroutine that will perform work
type Worker struct {
	Id            int
	TaskQueue     <-chan interface{}
	ResultChannel chan<- Result
}

// Result The struct that will be published to Worker.ResultChannel when workers complete their task.
type Result struct {
	WorkerId int
	Input    interface{}
	Data     interface{}
	Error    error
}

// Start Starts performing th task in goroutine
func (worker *Worker) Start(task Task) {
	go func() {
		for input := range worker.TaskQueue {
			result, err := task(input)

			// Publish the result of the work to the channel
			worker.ResultChannel <- Result{
				WorkerId: worker.Id,
				Input:    input,
				Data:     result,
				Error:    err,
			}
		}
	}()
}

// WorkerPool Manages the workers, handles tasks, and manages results
type WorkerPool struct {
	WorkerCount   int              // The amount of workers in the pool
	TaskQueue     chan interface{} // Channel that holds tasks for the pool to run
	ResultChannel chan Result      // Channel where the result of the tasks will be published
}

// NewWorkerPool Initializes a new worker pool
func NewWorkerPool(workerCount int) *WorkerPool {
	return &WorkerPool{
		WorkerCount:   workerCount,
		TaskQueue:     make(chan interface{}),
		ResultChannel: make(chan Result),
	}
}

// Start Starts the worker pool and all of its workers
func (pool *WorkerPool) Start(task Task) {
	for i := 0; i < pool.WorkerCount; i++ {
		worker := Worker{
			Id:            i,
			TaskQueue:     pool.TaskQueue,
			ResultChannel: pool.ResultChannel,
		}

		worker.Start(task)
	}
}

// AddTask Adds a task to the queue
func (pool *WorkerPool) AddTask(input interface{}) {
	pool.TaskQueue <- input
}

// GetResult Gets a result from the pool's channel
func (pool *WorkerPool) GetResult() Result {
	return <-pool.ResultChannel
}
