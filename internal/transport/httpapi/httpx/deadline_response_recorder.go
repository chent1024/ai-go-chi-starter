package httpx

import (
	"context"
	"net/http"
)

type DeadlineAwareResponseRecorder struct {
	*ResponseRecorder
	ctx            context.Context
	lateWriteCount int
}

func NewDeadlineAwareResponseRecorder(w http.ResponseWriter, ctx context.Context) *DeadlineAwareResponseRecorder {
	return &DeadlineAwareResponseRecorder{
		ResponseRecorder: NewResponseRecorder(w),
		ctx:              ctx,
	}
}

func (r *DeadlineAwareResponseRecorder) WriteHeader(statusCode int) {
	if r == nil {
		return
	}
	if deadlineExceeded(r.ctx) {
		r.lateWriteCount++
		return
	}
	r.ResponseRecorder.WriteHeader(statusCode)
}

func (r *DeadlineAwareResponseRecorder) Write(p []byte) (int, error) {
	if r == nil {
		return 0, context.DeadlineExceeded
	}
	if deadlineExceeded(r.ctx) {
		r.lateWriteCount++
		return 0, context.DeadlineExceeded
	}
	return r.ResponseRecorder.Write(p)
}

func (r *DeadlineAwareResponseRecorder) LateWriteCount() int {
	if r == nil {
		return 0
	}
	return r.lateWriteCount
}

func deadlineExceeded(ctx context.Context) bool {
	return ctx != nil && ctx.Err() == context.DeadlineExceeded
}
