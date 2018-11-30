package swagger

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/qjw/go-gin-binding"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"strings"
	"fmt"
)

type DocLoader func(key string) ([]byte, error)

type GroupSwaggerData map[string]SwaggerEntry
type options struct {
	// 是否调试模式
	debugFlag bool
	// doc文档定义路径
	docPath string
	// url前缀
	baseUrl string
	// 文件和里面内容的缓存
	swaggerData map[string]SwaggerEntry
	// 文件和里面内容的缓存--增加分组功能
	swaggerDataMap map[string]GroupSwaggerData
	// 文档对象缓存
	docData map[string]*SwaggerDocFile
	// 当不从文件中加载doc时，获取swagger数据的loader
	docLoader DocLoader
	// 当前的配置
	config *Config
}

func newOptions(config *Config) *options {
	opt := &options{
		debugFlag:   config.Debug,
		docPath:     config.DocFilePath,
		baseUrl:     config.BasePath,
		swaggerData: make(map[string]SwaggerEntry),
		swaggerDataMap: make(map[string]GroupSwaggerData),
		docData:     make(map[string]*SwaggerDocFile),
		config: config,
	}
	// 非调试模式，不允许从外部加载doc文件
	if !opt.debugFlag {
		opt.docPath = ""
	}
	if len(opt.docPath) > 0 {
		if path, err := filepath.Abs(opt.docPath); err != nil {
			panic(err)
			return nil
		} else {
			if strings.HasSuffix(path, "/") {
				opt.docPath = path
			} else {
				opt.docPath = path + "/"
			}
		}
	}
	return opt
}

// 使用分组的option
//var (
//	gOption *options = nil
//)
var (
	gDefaultOption *options = nil
	gGroupOptions = make(map[string]*options)
)

func swaggerImp(docFile *SwaggerDocFile, path string, method string, entry string) {
	if data, ok := (*docFile)[entry]; ok {
		swaggerFinish(path, method, &data)
	} else {
		panic(errors.New("file'" + entry + "' have no entry '" + entry + "'"))
	}
}

func parseFileNode(entry string) (filepath string, node string, err error) {
	entrys := strings.Split(entry, ":")
	if len(entrys) != 2 {
		err = errors.New("invalid swagger entry '" + entry + "'")
		return
	}

	filepath = entrys[0]
	node = entrys[1]
	if len(filepath) == 0 {
		err = errors.New("invalid swagger entry '" + entry + "',invalid filepath")
		return
	}
	if len(node) == 0 {
		err = errors.New("invalid swagger entry '" + entry + "',invalid node")
		return
	}
	return
}

func realPath(group *gin.RouterGroup, path string) string {
	var burl = group.BasePath()
	if len(gDefaultOption.baseUrl) > 0 && strings.HasPrefix(burl, gDefaultOption.baseUrl) {
		//burl = strings.TrimPrefix(burl, gDefaultOption.baseUrl)
	}
	if strings.HasPrefix(path, "/") {
		path = burl + path
	} else {
		path = burl + "/" + path
	}
	return path
}

func realPathByGroup(group *gin.RouterGroup, apiGroup, path string) string {
	var (
		option *options
		exist bool
	)
	option, exist = gGroupOptions[apiGroup]
	if !exist {
		panic(fmt.Errorf("Not find group option %s", apiGroup))
	}

	var burl = group.BasePath()
	if len(option.baseUrl) > 0 && strings.HasPrefix(burl, option.baseUrl) {
		//burl = strings.TrimPrefix(burl, option.baseUrl)
	}

	if strings.HasPrefix(path, "/") {
		path = burl + path
	} else {
		path = burl + "/" + path
	}

	return path
}

func Swagger2(group *gin.RouterGroup, path string, method string, extra *StructParam) {
	path = realPath(group, path)
	swaggerFinish(path, method, NewSwaggerMethodEntry(extra))
}

func Swagger2ByGroup(group *gin.RouterGroup, apiGroup string, path string, method string, extra *StructParam) {
	path = realPathByGroup(group, apiGroup, path)
	swaggerFinishByGroup(apiGroup, path, method, NewSwaggerMethodEntry(extra))
}

func Swagger(group *gin.RouterGroup, path string, method string, entry string) {
	path = realPath(group, path)
	// 解析文件路径和内部路径
	var err error
	rfilepath, node, err := parseFileNode(entry)
	if err != nil {
		panic(err)
	}

	// 是否有缓存
	if docFile, ok := gDefaultOption.docData[rfilepath]; ok {
		swaggerImp(docFile, path, method, node)
		return
	}

	var yamlFile []byte
	if len(gDefaultOption.docPath) > 0 {
		rfilepath = gDefaultOption.docPath + rfilepath

		// 加载文件
		yamlFile, err = ioutil.ReadFile(rfilepath)
		if err != nil {
			panic(err)
		}
	} else {
		yamlFile, err = gDefaultOption.docLoader(rfilepath)
		if err != nil {
			panic(err)
		}
	}

	var docFile SwaggerDocFile
	err = yaml.Unmarshal(yamlFile, &docFile)
	if err != nil {
		panic(err)
	}

	// 写入缓存
	gDefaultOption.docData[rfilepath] = &docFile
	swaggerImp(&docFile, path, method, node)
}

func swaggerFinish(path string, method string, entry *SwaggerMethodEntry) {
	if err := binding.Validate(entry); err != nil {
		//panic(err)
		//return
	}

	var sentry SwaggerEntry
	if v, ok := gDefaultOption.swaggerData[path]; ok {
		sentry = v
	} else {
		sentry = SwaggerEntry{}
	}
	sentry.SetMethod(method, *entry)
	gDefaultOption.swaggerData[path] = sentry
}

func swaggerFinishByGroup(apiGroup string, path string, method string, entry *SwaggerMethodEntry) {
	if err := binding.Validate(entry); err != nil {
		//panic(err)
		//return
	}

	var (
		option *options
		exist bool
	)
	option, exist = gGroupOptions[apiGroup]
	if !exist {
		panic(fmt.Errorf("Not find group option %s", apiGroup))
	}

	var sentry SwaggerEntry
	var groupData GroupSwaggerData
	if v, ok := option.swaggerDataMap[apiGroup]; ok {
		groupData = v
	} else {
		groupData = make(GroupSwaggerData)
	}

	if v, ok := groupData[path]; ok {
		sentry = v
	} else {
		sentry = SwaggerEntry{}
	}

	sentry.SetMethod(method, *entry)

	groupData[path] = sentry
	option.swaggerDataMap[apiGroup] = groupData
}

func SwaggerSection(group *gin.RouterGroup, apiGroup string, section string, content string) {
	var (
		option *options
		exist bool
	)

	option, exist = gGroupOptions[apiGroup]
	if !exist {
		option = gDefaultOption
	}

	sectionContent := "\n\n\n" + section + "\n==========\n\n\n" + content
	option.config.Description += sectionContent
}

//func GetFlags(baseUrl string) []cli.Flag {
//	return []cli.Flag{
//		cli.StringFlag{
//			EnvVar: "SWAGGER_BASE_PATH",
//			Name:   "swagger_base_path",
//			Usage:  "api前缀，例如/api/v1",
//			Value:  baseUrl,
//		},
//		cli.StringFlag{
//			EnvVar: "SWAGGER_VERSION",
//			Name:   "swagger_version",
//			Usage:  "swagger版本号",
//			Value:  "2.0",
//		},
//		cli.StringFlag{
//			EnvVar: "SWAGGER_DOC_TITLE",
//			Name:   "swagger_doc_title",
//			Usage:  "swagger文档标题",
//			Value:  "Swagger文档",
//		},
//		cli.StringFlag{
//			EnvVar: "SWAGGER_DOC_DESC",
//			Name:   "swagger_doc_desc",
//			Usage:  "swagger文档描述",
//			Value:  "Swagger文档手册",
//		},
//		cli.StringFlag{
//			EnvVar: "SWAGGER_DOC_VERSION",
//			Name:   "swagger_doc_version",
//			Usage:  "swagger文档版本",
//			Value:  "0.01",
//		},
//		cli.StringFlag{
//			EnvVar: "SWAGGER_URL_PREFIX",
//			Name:   "swagger_url_prefix",
//			Usage:  "swagger url路径前缀",
//			Value:  "doc",
//		},
//		cli.StringFlag{
//			EnvVar: "SWAGGER_UI_URL",
//			Name:   "swagger_ui_url",
//			Usage:  "swagger ui的地址",
//			Value:  "http://swagger.qiujinwu.com",
//		},
//		cli.StringFlag{
//			EnvVar: "SWAGGER_DOC_FILE_PATH",
//			Name:   "swagger_doc_file_path",
//			Usage:  "swagger文档的地址，用于调试，release直接打包到二进制里面",
//			Value:  "",
//		},
//	}
//}