package block



type Transaction struct {
	ID []byte
	Input []TxtInput
	Output []TxtOutput
}

type TxtOutput struct {
   Value int
   PubKey string
}
type TxtInput struct {
	ID []byte
	Out int
	Sig string
}