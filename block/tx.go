package block



type TxtOutput struct {
	Value  int
	PubKey string
}
type TxtInput struct {
	ID  []byte
	Out int
	Sig string
}

func (t *Transaction) IsCoinBase() bool {
	return len(t.Input) == 1 && len(t.Input[0].ID) == 0 && t.Input[0].Out == -1
}

func (in *TxtInput) CanUnLock(data string) bool {
	return in.Sig == data
}

func (out *TxtOutput) CanBeUnLocked(data string) bool {
	return out.PubKey == data
}