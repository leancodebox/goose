package fileopt

import (
	"os"
	"path/filepath"
	"strings"
)

// Put 将数据存入文件
func Put[DataType []byte | string](data DataType, to string) (err error) {
	err = os.WriteFile(to, []byte(data), 0644)
	return
}

func FileGetContents(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// FilePutContents file_put_contents
func FilePutContents[DType string | []byte](filename string, data DType, isAppend ...bool) error {
	if dir := filepath.Dir(filename); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	bData := []byte(data)
	needAppend := len(isAppend) > 0 && isAppend[0] == true
	if needAppend {
		fl, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer fl.Close()
		_, err = fl.Write(bData)
		return err
	} else {
		return os.WriteFile(filename, bData, 0644)
	}
}

func IsExist(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func IsExistOrCreate(path string, init string) error {
	if IsExist(path) {
		return nil
	}
	return FilePutContents(path, init)
}

func DirExistOrCreate(dirPath string) error {
	if IsExist(dirPath) {
		return nil
	} else {
		return os.MkdirAll(dirPath, 0755)
	}
}

func AbsPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") || path == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path, err
		}
		path = filepath.Join(homeDir, path[2:])
	}
	return path, nil
}

func FileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
