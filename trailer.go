package httparser

type trailerState int8

const (
	defaultTrailer trailerState = iota
	needParserTrailer
)
