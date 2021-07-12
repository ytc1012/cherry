package cherryDataConfig

import (
	cherryError "github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/extend/file"
	"github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/profile"
	"github.com/radovskyb/watcher"
	"io/ioutil"
	"time"
)

// SourceFile 本地读取数据配置文件
type SourceFile struct {
	monitorPath    string
	watcher        *watcher.Watcher
	reloadTime     int64
	extName        string
	configChangeFn ConfigChangeFn
}

func (f *SourceFile) Name() string {
	return "file"
}

func (f *SourceFile) Init(_ IDataConfig) {
	//read data_config->file node
	config := cherryProfile.GetConfig("data_config")
	fileNode := config.Get("file")
	if fileNode == nil {
		cherryLogger.Warnf("`data_config` node in `%s` file not found.", cherryProfile.FileName())
		return
	}

	filePath := fileNode.Get("file_path").ToString()
	if filePath == "" {
		//default value
		filePath = "data_config/"
	}

	f.extName = fileNode.Get("ext_name").ToString()
	if f.extName == "" {
		// default value
		f.extName = ".json"
	}

	var err error
	f.monitorPath, err = cherryFile.JoinPath(cherryProfile.Dir(), filePath)
	if err != nil {
		cherryLogger.Warn(err)
		return
	}

	f.reloadTime = fileNode.Get("reload_time").ToInt64()
	if f.reloadTime < 1 {
		//default value
		f.reloadTime = 3000
	}

	// new watcher
	go f.newWatcher()
}

func (f *SourceFile) ReadBytes(configName string) (data []byte, error error) {
	if configName == "" {
		return nil, cherryError.Error("configName is empty.")
	}

	fullPath, err := cherryFile.JoinPath(f.monitorPath, configName+f.extName)
	if err != nil {
		return nil, cherryError.Errorf("file not found. err = %v, fullPath = %s", err, fullPath)
	}

	if cherryFile.IsDir(fullPath) {
		return nil, cherryError.Errorf("path is dir. fullPath = %s", err, fullPath)
	}

	data, err = ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, cherryError.Errorf("read file err. err = %v path = %s", err, fullPath)
	}

	if len(data) < 1 {
		return nil, cherryError.Error("configName data is err.")
	}

	return data, nil
}

func (f *SourceFile) OnChange(fn ConfigChangeFn) {
	f.configChangeFn = fn
}

func (f *SourceFile) newWatcher() {
	f.watcher = watcher.New()
	f.watcher.SetMaxEvents(1)
	f.watcher.FilterOps(watcher.Write)

	err := f.watcher.Add(f.monitorPath)
	if err != nil {
		cherryLogger.Warn("new watcher error. path=%s, err=%v", f.monitorPath, err)
		return
	}

	//new goroutine
	go func() {
		for {
			select {
			case ev := <-f.watcher.Event:
				{
					if ev.IsDir() {
						return
					}

					configName := cherryFile.GetFileName(ev.FileInfo.Name(), true)
					cherryLogger.Infof("[name = %s] file change.", configName)

					data, err := f.ReadBytes(configName)
					if err != nil {
						cherryLogger.Error(err)
						return
					}

					if f.configChangeFn != nil {
						f.configChangeFn(configName, data)
					}
				}
			case err := <-f.watcher.Error:
				{
					cherryLogger.Error(err)
					return
				}
			case <-f.watcher.Closed:
				return
			}
		}
	}()

	if err := f.watcher.Start(time.Millisecond * time.Duration(f.reloadTime)); err != nil {
		cherryLogger.Warn(err)
	}
}

func (f *SourceFile) Stop() {
	if f.watcher != nil {
		err := f.watcher.Remove(f.monitorPath)
		if err != nil {
			cherryLogger.Warn(err)
		}
		cherryLogger.Infof("remove watcher [path = %s]", f.monitorPath)
		f.watcher.Closed <- struct{}{}
	}
}