package feature

import (
	"code.google.com/p/goprotobuf/proto"
	"io/ioutil"
)

type FeatureIndex struct {
	idx  map[string]uint32
	list *FeatureList
}

func New() *FeatureIndex {
	return &FeatureIndex{idx: map[string]uint32{}, list: &FeatureList{Features: []*Feature{}}}
}

func Load(filename string) (f *FeatureIndex, err error) {
	f = New()
	f.list = &FeatureList{Features: []*Feature{}}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(data, f.list)
	f.index()
	return
}

func (f *FeatureIndex) Save(filename string) (err error) {
	data, err := proto.Marshal(f.list)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(filename, data, 0644)
	return
}

func (f *FeatureIndex) Add(name string, index []uint32) {
	newf := &Feature{Name: proto.String(name), Index: []*Index{}}
	for _, i := range index {
		newf.Index = append(newf.Index, &Index{I: proto.Uint32(i)})
	}
	f.list.Features = append(f.list.Features, newf)
	f.idx[name] = uint32(len(f.list.Features) - 1)
	return
}

func (f *FeatureIndex) Find(name string) (index []uint32) {
	if i, has := f.idx[name]; has {
		for _, v := range f.list.Features[i].Index {
			index = append(index, *v.I)
		}
	}
	return
}

func (f *FeatureIndex) index() {
	for i, v := range f.list.Features {
		f.idx[*v.Name] = uint32(i)
	}
}
