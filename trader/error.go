package trader

func handleError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
