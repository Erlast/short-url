package storages

type Storage struct {
	Urls map[string]string
}

func Init(u map[string]string) Storage {
	return Storage{Urls: u}
}
