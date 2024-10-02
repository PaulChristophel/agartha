package model

type Tabler interface {
	TableName() string
}

type Marshaler interface {
	MarshalJSON() ([]byte, error)
}
