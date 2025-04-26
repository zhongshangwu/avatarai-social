package main

import (
	"reflect"

	"github.com/bluesky-social/indigo/mst"

	cbg "github.com/whyrusleeping/cbor-gen"

	"github.com/zhongshangwu/avatarai-social/pkg/atproto/vtri"
)

func main() {
	var typVals []any
	for _, typ := range mst.CBORTypes() {
		typVals = append(typVals, reflect.New(typ).Elem().Interface())
	}

	genCfg := cbg.Gen{
		MaxStringLength: 1_000_000,
	}

	if err := genCfg.WriteMapEncodersToFile("pkg/atproto/vtri/cbor_gen.go", "vtri",
		vtri.AvatarProfile{},
		vtri.AsterProfile{},
		vtri.EntityExternal{},
		vtri.EntityExternal_External{},
		vtri.EntityImages{},
		vtri.EntityImages_Image{},
		vtri.EntityVideo{},
		vtri.EntityVideo_Caption{},
		vtri.ActivityMoment_ReplyRef{},
		vtri.EntityRecord{},
		vtri.EntityDefs_AspectRatio{},
		vtri.ActivityMoment_Embed{},
		vtri.ActivityMoment{},
	); err != nil {
		panic(err)
	}
}
