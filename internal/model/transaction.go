package model

type Tx interface {
	Rollback() error
	Commit() error
}

type Transaction struct {
	tx Tx
	commited bool
}

func NewTransaction(tx Tx) (*Transaction, error) {
	return &Transaction{tx: tx, commited: false}, nil
}

func (t *Transaction) Rollback() error {
	if t.commited {
		return nil
	}
	return t.tx.Rollback()
}

func (t *Transaction) Commit() error {
	t.commited = true
	return t.tx.Commit()
}
