package adminbase

import (
	"errors"
	"fmt"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"github.com/lennon-guan/pipe"
	"github.com/ziipin-server/niuhe"
	"net/http"
	"reflect"
	"strconv"
	"sync"
)

type Any map[string]interface{}

type ColRenderer struct {
	ColName    string
	RenderFunc func(interface{}) interface{}
}

func RenderField(fieldName string) func(interface{}) interface{} {
	fieldNameValue := reflect.ValueOf(fieldName)
	return func(item interface{}) interface{} {
		itemVal := reflect.ValueOf(item)
		itemTyp := itemVal.Type()
		var fieldVal reflect.Value
		if itemTyp.Kind() == reflect.Struct {
			fieldVal = itemVal.FieldByName(fieldName)
		} else if itemTyp.Kind() == reflect.Map {
			fieldVal = itemVal.MapIndex(fieldNameValue)
		}
		return fieldVal.Interface()
	}
}

type FieldMapping struct {
	FormName     string
	ModelName    string
	ToFormValue  func(interface{}) interface{}
	ToModelValue func(string) (interface{}, error)
}

type AdminCrudViewCtrl struct {
	modelType  reflect.Type
	GetEngine  func() *xorm.Engine
	dbInitOnce sync.Once
	// Common fields
	pkList       []string
	pkArgList    []string
	autoIncrList []string
	colMap       map[string]*core.Column
	// List Ctrl fields
	ColRenderers []ColRenderer
	FilterList   func(c *niuhe.Context, session *xorm.Session) *xorm.Session
	// Edit Ctrl fields
	EditFormFieldMappings []FieldMapping
	BeforeEditSave        func(c *niuhe.Context, model interface{}, session *xorm.Session) error
	EditSaved             func(c *niuhe.Context, model interface{}, session *xorm.Session) error
	editing               struct {
		toFormMap  map[string]*FieldMapping
		toModelMap map[string]*FieldMapping
		updateCols []string
	}
	// Add Ctrl fields
	AddFormFieldMappings []FieldMapping
	BeforeAddSave        func(c *niuhe.Context, model interface{}, session *xorm.Session) error
	AddSaved             func(c *niuhe.Context, model interface{}, session *xorm.Session) error
	adding               struct {
		toFormMap  map[string]*FieldMapping
		toModelMap map[string]*FieldMapping
		updateCols []string
	}
}

func (ctrl *AdminCrudViewCtrl) Init(newModel interface{}) {
	// Init model type
	modelType := reflect.TypeOf(newModel)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	ctrl.modelType = modelType
}

func (ctrl *AdminCrudViewCtrl) initDbInfo() {
	ctrl.dbInitOnce.Do(func() {
		engine := ctrl.GetEngine()
		tableInfo := engine.TableInfo(reflect.New(ctrl.modelType).Interface())
		columns := tableInfo.Columns()
		ctrl.colMap = pipe.NewPipe(columns).
			ToMap(
			func(col *core.Column) string { return col.FieldName },
			nil,
		).(map[string]*core.Column)
		ctrl.pkList = pipe.NewPipe(columns).
			Filter(func(col *core.Column) bool { return col.IsPrimaryKey }).
			Map(func(col *core.Column) string { return col.FieldName }).
			ToSlice().([]string)
		ctrl.pkArgList = pipe.NewPipe(columns).
			Filter(func(col *core.Column) bool { return col.IsPrimaryKey }).
			Map(func(col *core.Column) string { return col.Name }).
			ToSlice().([]string)
		ctrl.autoIncrList = pipe.NewPipe(columns).
			Filter(func(col *core.Column) bool { return col.IsAutoIncrement }).
			Map(func(col *core.Column) string { return col.FieldName }).
			ToSlice().([]string)
		// Init editing
		ctrl.editing.updateCols = make([]string, 0)
		if ctrl.EditFormFieldMappings != nil {
			ctrl.editing.toFormMap = make(map[string]*FieldMapping, len(ctrl.EditFormFieldMappings))
			ctrl.editing.toModelMap = make(map[string]*FieldMapping, len(ctrl.EditFormFieldMappings))
			for i := 0; i < len(ctrl.EditFormFieldMappings); i++ {
				mapping := &ctrl.EditFormFieldMappings[i]
				ctrl.editing.toFormMap[mapping.ModelName] = mapping
				ctrl.editing.toModelMap[mapping.FormName] = mapping
				ctrl.editing.updateCols = append(ctrl.editing.updateCols, ctrl.colMap[mapping.ModelName].Name)
			}
		} else {
			fieldNum := ctrl.modelType.NumField()
			ctrl.editing.toFormMap = make(map[string]*FieldMapping, fieldNum)
			ctrl.editing.toModelMap = make(map[string]*FieldMapping, fieldNum)
			for i := 0; i < fieldNum; i++ {
				field := ctrl.modelType.Field(i)
				if col, exists := ctrl.colMap[field.Name]; exists && col.IsPrimaryKey {
					continue
				}
				ctrl.editing.updateCols = append(ctrl.editing.updateCols, ctrl.colMap[field.Name].Name)
				mapping := &FieldMapping{field.Name, field.Name, nil, nil}
				ctrl.editing.toFormMap[field.Name] = mapping
				ctrl.editing.toModelMap[field.Name] = mapping
			}
		}
		// Init adding
		if ctrl.AddFormFieldMappings != nil {
			ctrl.adding.toFormMap = make(map[string]*FieldMapping, len(ctrl.AddFormFieldMappings))
			ctrl.adding.toModelMap = make(map[string]*FieldMapping, len(ctrl.AddFormFieldMappings))
			for i := 0; i < len(ctrl.AddFormFieldMappings); i++ {
				mapping := &ctrl.AddFormFieldMappings[i]
				ctrl.adding.toFormMap[mapping.ModelName] = mapping
				ctrl.adding.toModelMap[mapping.FormName] = mapping
				ctrl.adding.updateCols = append(ctrl.adding.updateCols, ctrl.colMap[mapping.ModelName].Name)
			}
		} else {
			fieldNum := ctrl.modelType.NumField()
			ctrl.adding.toFormMap = make(map[string]*FieldMapping, fieldNum)
			ctrl.adding.toModelMap = make(map[string]*FieldMapping, fieldNum)
			for i := 0; i < fieldNum; i++ {
				field := ctrl.modelType.Field(i)
				if col, exists := ctrl.colMap[field.Name]; exists && col.IsAutoIncrement {
					continue
				}
				mapping := &FieldMapping{field.Name, field.Name, nil, nil}
				ctrl.adding.toFormMap[field.Name] = mapping
				ctrl.adding.toModelMap[field.Name] = mapping
			}
		}
	})
}

func (ctrl *AdminCrudViewCtrl) ReturnOK(c *niuhe.Context, datas interface{}) {
	c.JSON(http.StatusOK, Any{
		"result": 0,
		"data":   datas,
	})
}

func (ctrl *AdminCrudViewCtrl) ReturnError(c *niuhe.Context, err error) {
	if ierr, isCommErr := err.(niuhe.ICommError); isCommErr {
		c.JSON(http.StatusOK, Any{
			"result":  ierr.GetCode(),
			"message": ierr.GetMessage(),
		})
	} else {
		c.JSON(http.StatusInternalServerError, Any{
			"result":  -1,
			"message": err.Error(),
		})
	}
}

func (ctrl *AdminCrudViewCtrl) newSession() *xorm.Session {
	ctrl.initDbInfo()
	return ctrl.GetEngine().NewSession()
}

func (ctrl *AdminCrudViewCtrl) ToRow(item interface{}) (res map[string]interface{}) {
	itemVal := reflect.ValueOf(item)
	itemTyp := itemVal.Type()
	res = make(map[string]interface{})
	if ctrl.ColRenderers == nil {
		if itemTyp.Kind() == reflect.Struct {
			numField := itemTyp.NumField()
			for i := 0; i < numField; i++ {
				res[itemTyp.Field(i).Name] = itemVal.Field(i).Interface()
			}
		} else if itemTyp.Kind() == reflect.Map {
			keys := itemVal.MapKeys()
			for _, key := range keys {
				res[key.String()] = itemVal.MapIndex(key).Interface()
			}
		}
	} else {
		for _, renderer := range ctrl.ColRenderers {
			res[renderer.ColName] = renderer.RenderFunc(item)
		}
	}
	return
}

func (ctrl *AdminCrudViewCtrl) ApplyFilterList(c *niuhe.Context, session *xorm.Session) *xorm.Session {
	if ctrl.FilterList != nil {
		return ctrl.FilterList(c, session)
	} else {
		return session
	}
}

func (ctrl *AdminCrudViewCtrl) GetPage(c *niuhe.Context) (outPage int, outPageSize int, outTotal int, outRows []map[string]interface{}, err error) {
	outPage = 1
	outPageSize = 10
	if s, find := c.GetQuery("page"); find {
		outPage, err = strconv.Atoi(s)
		if err != nil || outPage < 1 {
			outPage = 1
		}
	}
	if s, find := c.GetQuery("page_size"); find {
		outPageSize, err = strconv.Atoi(s)
		if err != nil {
			outPageSize = 10
		} else if outPageSize < 1 {
			outPageSize = 1
		} else if outPageSize > 100 {
			outPageSize = 100
		}
	}
	err = nil
	dbSession := ctrl.newSession()
	defer dbSession.Close()
	// get total
	dbSession = ctrl.ApplyFilterList(c, dbSession)
	var outTotal64 int64
	if outTotal64, err = dbSession.Count(reflect.New(ctrl.modelType).Interface()); err != nil {
		return
	}
	outTotal = int(outTotal64)
	// get rows
	dbSession = ctrl.ApplyFilterList(c, dbSession)
	mSlice := reflect.MakeSlice(reflect.SliceOf(ctrl.modelType), 0, outPageSize)
	mSlicePtr := reflect.New(mSlice.Type())
	mSlicePtr.Elem().Set(mSlice)
	offset := (outPage - 1) * outPageSize
	if err = dbSession.Limit(outPageSize, offset).Find(mSlicePtr.Interface()); err != nil {
		return
	} else {
		outRows = pipe.NewPipe(mSlicePtr.Elem().Interface()).
			Map(ctrl.ToRow).
			ToSlice().([]map[string]interface{})
	}
	return
}

func (ctrl *AdminCrudViewCtrl) getPks(c *niuhe.Context, getOnly bool) (core.PK, error) {
	pks := core.PK{}
	var argVal string
	var exists bool
	for _, argName := range ctrl.pkArgList {
		if getOnly {
			argVal, exists = c.GetQuery(argName)
		} else {
			argVal, exists = c.GetPostForm(argName)
		}
		if !exists {
			return nil, errors.New("no " + argName)
		}
		pks = append(pks, argVal)
	}
	return pks, nil
}

func (ctrl *AdminCrudViewCtrl) getModelByPks(pks core.PK, session *xorm.Session) (interface{}, error) {
	model := reflect.New(ctrl.modelType).Interface()
	exists, err := session.Id(pks).Get(model)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}
	return model, nil
}

func (ctrl *AdminCrudViewCtrl) getModelByCtx(c *niuhe.Context, session *xorm.Session, getOnly bool) (interface{}, error) {
	pks, err := ctrl.getPks(c, getOnly)
	if err != nil {
		return nil, err
	}
	return ctrl.getModelByPks(pks, session)
}

func (ctrl *AdminCrudViewCtrl) ToForm(model interface{}, mappings map[string]*FieldMapping) interface{} {
	numField := ctrl.modelType.NumField()
	res := make(map[string]interface{})
	modelVal := reflect.ValueOf(model)
	if modelVal.Type().Kind() == reflect.Ptr {
		modelVal = modelVal.Elem()
	}
	for i := 0; i < numField; i++ {
		fieldInfo := ctrl.modelType.Field(i)
		if mapping, exists := mappings[fieldInfo.Name]; exists {
			if mapping.ToFormValue != nil {
				res[mapping.FormName] = mapping.ToFormValue(modelVal.Field(i).Interface())
			} else {
				res[mapping.FormName] = modelVal.Field(i).Interface()
			}
		}
	}
	return res
}

func (ctrl *AdminCrudViewCtrl) setFieldValue(fieldVal reflect.Value, fieldTyp reflect.StructField, valStr string) error {
	bitsize := int(fieldTyp.Type.Size()) * 8
	fieldName := fieldTyp.Name
	switch fieldTyp.Type.Kind() {
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		if ival, err := strconv.ParseInt(valStr, 10, bitsize); err != nil {
			niuhe.LogError("mergeModel fail to set field %s value %v", fieldName, valStr)
			return err
		} else {
			fieldVal.SetInt(ival)
		}
		break
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		if uval, err := strconv.ParseUint(valStr, 10, bitsize); err != nil {
			niuhe.LogError("mergeModel fail to set field %s value %v", fieldName, valStr)
			return err
		} else {
			fieldVal.SetUint(uval)
		}
		break
	case reflect.String:
		fieldVal.SetString(valStr)
		break
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		if fval, err := strconv.ParseFloat(valStr, bitsize); err != nil {
			niuhe.LogError("mergeModel fail to set field %s value %v", fieldName, valStr)
			return err
		} else {
			fieldVal.SetFloat(fval)
		}
		break
	case reflect.Bool:
		if ival, err := strconv.Atoi(valStr); err == nil {
			fieldVal.SetBool(ival != 0)
		} else if bval, err := strconv.ParseBool(valStr); err != nil {
			niuhe.LogError("mergeModel fail to set field %s value %v", fieldName, valStr)
			return err
		} else {
			fieldVal.SetBool(bval)
		}
		break
	default:
		return errors.New(fmt.Sprintf("mergeModel fail: no support field %s kind %s",
			fieldName, fieldTyp.Type.Kind().String()))
	}
	return nil
}

func (ctrl *AdminCrudViewCtrl) mergeModel(c *niuhe.Context, model interface{}, mappingMap map[string]*FieldMapping) error {
	modelValue := reflect.ValueOf(model)
	if modelValue.Type().Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
	}
	for formName, mapping := range mappingMap {
		valStr := c.PostForm(formName)
		fieldVal := modelValue.FieldByName(mapping.ModelName)
		if !fieldVal.IsValid() {
			return errors.New("Cannot find field value " + mapping.ModelName)
		}
		fieldTyp, exists := ctrl.modelType.FieldByName(mapping.ModelName)
		if !exists {
			return errors.New("Cannot find field type " + mapping.ModelName)
		}
		if mapping.ToModelValue != nil {
			if mValue, err := mapping.ToModelValue(valStr); err != nil {
				return err
			} else {
				fieldVal.Set(reflect.ValueOf(mValue))
			}
		} else if err := ctrl.setFieldValue(fieldVal, fieldTyp, valStr); err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *AdminCrudViewCtrl) GetEditModel(c *niuhe.Context) (interface{}, error) {
	session := ctrl.newSession()
	defer session.Close()
	model, err := ctrl.getModelByCtx(c, session, true)
	if err != nil {
		return nil, err
	} else if model == nil {
		return nil, nil
	}
	return ctrl.ToForm(model, ctrl.editing.toFormMap), nil
}

func (ctrl *AdminCrudViewCtrl) SaveEditModel(c *niuhe.Context) error {
	session := ctrl.newSession()
	defer session.Close()
	pks, err := ctrl.getPks(c, true)
	if err != nil {
		return err
	}
	model, err := ctrl.getModelByPks(pks, session)
	if err != nil {
		return err
	}
	if model == nil {
		return errors.New("no this model")
	}
	if err := session.Begin(); err != nil {
		return err
	}
	if ctrl.BeforeEditSave != nil {
		if err := ctrl.BeforeEditSave(c, model, session); err != nil {
			return err
		}
	}
	if err := ctrl.mergeModel(c, model, ctrl.editing.toModelMap); err != nil {
		return err
	}
	if _, err := session.Id(pks).Cols(ctrl.editing.updateCols...).Update(model); err != nil {
		return err
	}
	if ctrl.EditSaved != nil {
		if err := ctrl.EditSaved(c, model, session); err != nil {
			return err
		}
	}
	if err := session.Commit(); err != nil {
		return err
	}
	return nil
}

func (ctrl *AdminCrudViewCtrl) GetAddModel(c *niuhe.Context) (interface{}, error) {
	session := ctrl.newSession()
	defer session.Close()
	model := reflect.New(ctrl.modelType).Interface()
	return ctrl.ToForm(model, ctrl.editing.toFormMap), nil
}

func (ctrl *AdminCrudViewCtrl) SaveAddModel(c *niuhe.Context) error {
	session := ctrl.newSession()
	defer session.Close()
	model := reflect.New(ctrl.modelType).Interface()
	if err := session.Begin(); err != nil {
		return err
	}
	if ctrl.BeforeAddSave != nil {
		if err := ctrl.BeforeAddSave(c, model, session); err != nil {
			return err
		}
	}
	if err := ctrl.mergeModel(c, model, ctrl.adding.toModelMap); err != nil {
		return err
	}
	if affected, err := session.Insert(model); err != nil {
		return err
	} else if affected < 1 {
		return errors.New("Insert affected 0")
	}
	if ctrl.AddSaved != nil {
		if err := ctrl.AddSaved(c, model, session); err != nil {
			return err
		}
	}
	if err := session.Commit(); err != nil {
		return err
	}
	return nil
}

func (ctrl *AdminCrudViewCtrl) Del(c *niuhe.Context) error {
	session := ctrl.newSession()
	defer session.Close()
	pks, err := ctrl.getPks(c, false)
	if err != nil {
		return err
	}
	model := reflect.New(ctrl.modelType).Interface()
	if _, err := session.Id(pks).Delete(model); err != nil {
		return err
	}
	return nil
}
