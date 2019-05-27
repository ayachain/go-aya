package aapp

type mrg struct {
	aapps map[string]*aapp
}

func ( m *mrg ) Load( aappns string ) ( ap *aapp, err error ) {
	return &aapp{}, nil
}

