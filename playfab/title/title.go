package title

type Title string

func (t Title) URL(path string) string {
	return "https://" + t.String() + ".playfabapi.com" + path
}

func (t Title) String() string { return string(t) }
