package test

type DBMessageRecord struct {
	MessageId string
}

type DBActions struct {
	Res interface{}
}

func (r *DBActions) Execute() error {

	switch v := r.Res.(type) {
	case *DBMessageRecord:
		v.MessageId = "12345" // modify the original struct
	}

	return nil
}

// func main() {
// 	data := DBMessageRecord{}
// 	A := DBActions{
// 		Res: &data,
// 	}

// 	_ = A.Execute()

// 	fmt.Printf("%+v", data)
// }
