package bongo

type Tag struct {
	Name  string
	Pages PageList
}

type TagList []*Tag

func (t TagList) Len() int { return len(t) }
func (t TagList) Less(i, j int) bool {
	return t[i].Name < t[j].Name
}
func (t TagList) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
