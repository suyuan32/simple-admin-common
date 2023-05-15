package errormsg

const (
	Success          string = "successful"
	Fail             string = "failed"
	CreateSuccess    string = "create successfully"
	CreateFail       string = "create failed"
	UpdateSuccess    string = "update successfully"
	UpdateFail       string = "update failed"
	DeleteSuccess    string = "delete successfully"
	DeleteFail       string = "delete successfully"
	TargetNotFound   string = "target does not exist"
	ConstraintError  string = "data conflict"
	ValidationError  string = "validate failed"
	NotSingularError string = "data is not unique"
	DatabaseError    string = "database error"
	RedisError       string = "redis error"
)
