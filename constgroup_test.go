package niuhe

import "testing"

type LangEnumType struct {
	*IntConstGroup
	ZH_CN IntConstItem `value:"1" name:"简体中文"`
	WEIYU IntConstItem `value:"2" name:"维语"`
}

var LangEnum LangEnumType

func TestIntConstGroup(t *testing.T) {
	if LangEnum.ZH_CN.Value != 1 {
		t.Error("ZH_CN.Value error")
		return
	}
	if LangEnum.ZH_CN.Name != "简体中文" {
		t.Error("ZH_CN.Name error")
		return
	}
	if LangEnum.MustGetName(LangEnum.ZH_CN.Value) != LangEnum.ZH_CN.Name {
		t.Error("GetName error, returns " + LangEnum.GetName(LangEnum.ZH_CN.Value))
		return
	}
}

func init() {
	InitConstGroup(&LangEnum)
}
