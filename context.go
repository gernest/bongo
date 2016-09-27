package bongo

import "sort"

//Context contains the tree of processed pages.
type Context struct {
	Pages   PageList
	Tags    TagList
	Untaged PageList
	Data    map[string]interface{}
}

//GetAllSections retruns a Context object for the PageList. Thismakes surepages
//are arranged by tags,  pages with no tags are assigned to the Cotext.Untagged.
func GetContext(p PageList) *Context {
	ctx := &Context{Pages: p}
	for i := 0; i < len(p); i++ {
		pg := p[i]
		for _, t := range pg.Tags {
			if len(ctx.Tags) > 0 {
				sort.Sort(ctx.Tags)
				key := sort.Search(len(ctx.Tags), func(x int) bool {
					return ctx.Tags[x].Name >= t
				})
				if key != len(ctx.Tags) {
					ctx.Tags[key].Pages = append(ctx.Tags[key].Pages, pg)
				}
				var pl PageList
				pl = append(pl, pg)
				ctx.Tags = append(ctx.Tags, &Tag{Name: t, Pages: pl})
			} else {
				ctx.Untaged = append(ctx.Untaged, pg)
			}
		}
	}
	return ctx
}
