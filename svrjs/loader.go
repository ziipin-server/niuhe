package svrjs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type loaderItem struct {
	info   os.FileInfo
	script string
}

//ScriptLoader 脚本加载器
type ScriptLoader struct {
	debugging bool
	paths     []string
	idmap     map[string]string
	idrw      sync.RWMutex
	scripts   map[string]*loaderItem
	scriptsrw sync.RWMutex
}

//NewScriptLoader create a new script loader
func NewScriptLoader(paths []string, debugging bool) *ScriptLoader {
	return &ScriptLoader{
		debugging: debugging,
		paths:     paths,
		idmap:     make(map[string]string),
		scripts:   make(map[string]*loaderItem),
	}
}

func (l *ScriptLoader) loadScriptFromFileSystem(sp string) (string, error) {
	srcBytes, err := ioutil.ReadFile(sp)
	if err != nil {
		return "", err
	}
	src := "(function(exports, require, module, __filename, __dirname){ " + string(srcBytes) + "\n})"
	return src, nil
}

func (l *ScriptLoader) isFileExists(elem ...string) (string, bool) {
	ap, err := filepath.Abs(path.Join(elem...))
	if err != nil {
		return "", false
	}
	fi, err := os.Stat(ap)
	return ap, (err == nil || os.IsExist(err)) && !fi.IsDir()
}

//GetModuleAbs ...
func (l *ScriptLoader) GetModuleAbs(cwd, id string) (string, error) {
	var searchPaths []string
	var idkey string
	if strings.HasPrefix(id, ".") {
		searchPaths = []string{cwd}
		idkey = cwd + id
	} else {
		searchPaths = l.paths
		idkey = id
	}
	l.idrw.RLock()
	if ap, found := l.idmap[idkey]; found {
		l.idrw.RUnlock()
		return ap, nil
	}
	l.idrw.RUnlock()
	for _, searchPath := range searchPaths {
		if mp, found := l.isFileExists(searchPath, id); found {
			l.idrw.Lock()
			l.idmap[idkey] = mp
			l.idrw.Unlock()
			return mp, nil
		}
		if mp, found := l.isFileExists(searchPath, id+".js"); found {
			l.idrw.Lock()
			l.idmap[idkey] = mp
			l.idrw.Unlock()
			return mp, nil
		}
		if mp, found := l.isFileExists(searchPath, id, "index.js"); found {
			l.idrw.Lock()
			l.idmap[idkey] = mp
			l.idrw.Unlock()
			return mp, nil
		}
	}
	return "", fmt.Errorf("cannot find id %s in cwd %s", id, cwd)
}

//LoadScript load a script
func (l *ScriptLoader) LoadScript(sp string) (string, error) {
	l.scriptsrw.RLock()
	var newInfo os.FileInfo
	var err error
	if item, found := l.scripts[sp]; found {
		if !l.debugging {
			l.scriptsrw.RUnlock()
			return item.script, nil
		}
		if newInfo, err = os.Stat(sp); err != nil {
			l.scriptsrw.RUnlock()
			return "", err
		} else if !newInfo.ModTime().After(item.info.ModTime()) {
			l.scriptsrw.RUnlock()
			return item.script, nil
		}
	}
	if script, err := l.loadScriptFromFileSystem(sp); err != nil {
		l.scriptsrw.RUnlock()
		return "", nil
	} else {
		l.scriptsrw.RUnlock()
		l.scriptsrw.Lock()
		if newInfo == nil {
			newInfo, _ = os.Stat(sp)
		}
		l.scripts[sp] = &loaderItem{
			info:   newInfo,
			script: script,
		}
		l.scriptsrw.Unlock()
		return script, nil
	}
}
