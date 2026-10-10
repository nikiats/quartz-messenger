package usecase

type IDGenerator interface {
	Generate() int64
}
