package ziggurat

import "context"

type StartFunction func(ctx context.Context, routeNames []string)
type StopFunction func()

const ProcessingSuccess ProcessStatus = 0
const RetryMessage ProcessStatus = 1
const SkipMessage ProcessStatus = 2

type ProcessStatus int
