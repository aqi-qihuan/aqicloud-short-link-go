package enums

type TaskState string

const (
	TASK_LOCK    TaskState = "LOCK"
	TASK_FINISH  TaskState = "FINISH"
	TASK_CANCEL  TaskState = "CANCEL"
)
