package fileopt

import (
	"os"
	"path/filepath"
	"strings"
)

var basePath string

func SetBasePath(path string) {
	basePath = path
}

func GetStoragePath(filename string) string {
	return filepath.Join(basePath, filename)
}

func StorageGetE(filename string) (string,error) {
	data, err := FileGetContents(GetStoragePath(filename))
	return string(data),err
}

func StorageGet(filename string) string {
	data, _ := StorageGetE(filename)
	return string(data)
}

func StoragePut[DType string | []byte](filename string, data DType, append bool) error {
	return FilePutContents(basePath+filename, data, append)
}

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
	needAppend := false
	if len(isAppend) > 0 && isAppend[0] == true {
		needAppend = true
	}
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

func PutContent[DType string | []byte](filename string, data DType) {
	_ = FilePutContents(filename, data, false)
}

func AppendPutContent[DType string | []byte](filename string, data DType) {
	_ = FilePutContents(filename, data, true)
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

func IsExistOrCreate(path string, init string) bool {
	if IsExist(path) {
		return true
	}
	PutContent(path, init)
	return true
}

func DirExistOrCreate(dirPath string) bool {
	if IsExist(dirPath) {
		return true
	} else {
		return os.MkdirAll(dirPath, os.ModePerm) != nil
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
