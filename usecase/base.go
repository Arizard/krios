package usecase

// BooleanUseCase defines a use case which returns a boolean response to
// the response
type BooleanUseCase interface {
	Execute(h BooleanResponseHandler)
}

// BooleanResponseHandler defines a response handler which takes a boolean
// from a use case.
type BooleanResponseHandler interface {
	Handle(b bool)
	Response() bool
}
